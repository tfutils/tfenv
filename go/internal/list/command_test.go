package list

import (
	"bytes"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"encoding/binary"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- Test helpers for creating minimal binary files ---

// createTestELF writes a minimal valid ELF64 binary with the given machine type.
func createTestELF(t *testing.T, path string, machine elf.Machine) {
	t.Helper()
	hdr := elf.Header64{
		Ident: [elf.EI_NIDENT]byte{
			0: 0x7f, 1: 'E', 2: 'L', 3: 'F',
			4: 2, // ELFCLASS64
			5: 1, // ELFDATA2LSB
			6: 1, // EV_CURRENT
		},
		Type:    2, // ET_EXEC
		Machine: uint16(machine),
		Version: 1,
		Ehsize:  64,
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := binary.Write(f, binary.LittleEndian, &hdr); err != nil {
		t.Fatal(err)
	}
}

// testMachOHeader64 is the raw layout of a 64-bit Mach-O header.
type testMachOHeader64 struct {
	Magic    uint32
	CpuType  uint32
	SubType  uint32
	FileType uint32
	NCmds    uint32
	CmdsSize uint32
	Flags    uint32
	Reserved uint32
}

// createTestMachO writes a minimal valid 64-bit little-endian Mach-O binary.
func createTestMachO(t *testing.T, path string, cpu macho.Cpu) {
	t.Helper()
	hdr := testMachOHeader64{
		Magic:    0xFEEDFACF, // MH_MAGIC_64
		CpuType:  uint32(cpu),
		SubType:  3, // CPU_SUBTYPE_ALL
		FileType: 2, // MH_EXECUTE
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := binary.Write(f, binary.LittleEndian, &hdr); err != nil {
		t.Fatal(err)
	}
}

// createTestPE writes a minimal valid PE binary with the given machine type.
func createTestPE(t *testing.T, path string, machine uint16) {
	t.Helper()
	// Minimal PE: DOS header (96 bytes) + PE signature (4) + COFF header (20).
	peOffset := uint32(96)
	size := int(peOffset) + 4 + 20
	buf := make([]byte, size)

	// DOS magic.
	buf[0] = 'M'
	buf[1] = 'Z'

	// PE offset at 0x3C.
	binary.LittleEndian.PutUint32(buf[0x3C:], peOffset)

	// PE signature.
	copy(buf[peOffset:], []byte{'P', 'E', 0, 0})

	// COFF Machine field (first 2 bytes of COFF header).
	binary.LittleEndian.PutUint16(buf[peOffset+4:], machine)

	if err := os.WriteFile(path, buf, 0o755); err != nil {
		t.Fatal(err)
	}
}

// captureStdout redirects os.Stdout to a pipe, runs fn, and returns
// whatever was written. Not safe for use with t.Parallel.
func captureStdout(fn func()) string {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stdout = w
	fn()
	w.Close()
	os.Stdout = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// captureStderr redirects os.Stderr to a pipe, runs fn, and returns
// whatever was written. Not safe for use with t.Parallel.
func captureStderr(fn func()) string {
	old := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}
	os.Stderr = w
	fn()
	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// --- detectBinaryArch tests ---

func TestDetectBinaryArch_ELF(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		machine elf.Machine
		want    string
	}{
		{"amd64", elf.EM_X86_64, "amd64"},
		{"arm64", elf.EM_AARCH64, "arm64"},
		{"386", elf.EM_386, "386"},
		{"arm", elf.EM_ARM, "arm"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "binary")
			createTestELF(t, path, tc.machine)
			got := detectBinaryArch(path)
			if got != tc.want {
				t.Errorf("detectBinaryArch() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetectBinaryArch_MachO(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cpu  macho.Cpu
		want string
	}{
		{"amd64", macho.CpuAmd64, "amd64"},
		{"arm64", macho.CpuArm64, "arm64"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "binary")
			createTestMachO(t, path, tc.cpu)
			got := detectBinaryArch(path)
			if got != tc.want {
				t.Errorf("detectBinaryArch() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetectBinaryArch_PE(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		machine uint16
		want    string
	}{
		{"amd64", pe.IMAGE_FILE_MACHINE_AMD64, "amd64"},
		{"arm64", pe.IMAGE_FILE_MACHINE_ARM64, "arm64"},
		{"386", pe.IMAGE_FILE_MACHINE_I386, "386"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "binary")
			createTestPE(t, path, tc.machine)
			got := detectBinaryArch(path)
			if got != tc.want {
				t.Errorf("detectBinaryArch() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestDetectBinaryArch_Nonexistent(t *testing.T) {
	t.Parallel()
	got := detectBinaryArch("/nonexistent/path/binary")
	if got != "unknown" {
		t.Errorf("detectBinaryArch(nonexistent) = %q, want %q", got, "unknown")
	}
}

func TestDetectBinaryArch_InvalidFile(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "binary")
	if err := os.WriteFile(path, []byte("not a binary"), 0o755); err != nil {
		t.Fatal(err)
	}
	got := detectBinaryArch(path)
	if got != "unknown" {
		t.Errorf("detectBinaryArch(invalid) = %q, want %q", got, "unknown")
	}
}

// --- RunList tests ---

func TestRunList_MultipleVersions(t *testing.T) {
	dir := t.TempDir()
	versionsDir := filepath.Join(dir, "versions")

	for _, v := range []string{"1.5.0", "1.6.0", "1.4.0"} {
		vDir := filepath.Join(versionsDir, v)
		if err := os.MkdirAll(vDir, 0o755); err != nil {
			t.Fatal(err)
		}
		createTestELF(t, filepath.Join(vDir, "terraform"), elf.EM_X86_64)
	}

	t.Setenv("TFENV_CONFIG_DIR", dir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "1.5.0")

	var code int
	stdout := captureStdout(func() {
		code = RunList(nil)
	})
	if code != 0 {
		t.Fatalf("RunList() returned %d, want 0", code)
	}

	lines := strings.Split(strings.TrimRight(stdout, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), stdout)
	}

	// Sorted newest first: 1.6.0, 1.5.0 (active), 1.4.0.
	if !strings.HasPrefix(lines[0], "  1.6.0 (amd64)") {
		t.Errorf("line 0 = %q, want prefix '  1.6.0 (amd64)'", lines[0])
	}
	if !strings.HasPrefix(lines[1], "* 1.5.0 (amd64)") {
		t.Errorf("line 1 = %q, want prefix '* 1.5.0 (amd64)'", lines[1])
	}
	if !strings.Contains(lines[1], "(set by TFENV_TERRAFORM_VERSION)") {
		t.Errorf("line 1 = %q, want to contain '(set by TFENV_TERRAFORM_VERSION)'", lines[1])
	}
	if !strings.HasPrefix(lines[2], "  1.4.0 (amd64)") {
		t.Errorf("line 2 = %q, want prefix '  1.4.0 (amd64)'", lines[2])
	}
}

func TestRunList_NoVersions(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TFENV_CONFIG_DIR", dir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "")
	t.Setenv("TFENV_DIR", dir)

	var code int
	stderr := captureStderr(func() {
		code = RunList(nil)
	})
	if code != 1 {
		t.Errorf("RunList() returned %d, want 1", code)
	}
	if !strings.Contains(stderr, "No versions available") {
		t.Errorf("stderr = %q, want to contain 'No versions available'", stderr)
	}
}

func TestRunList_NoDefault(t *testing.T) {
	dir := t.TempDir()
	versionsDir := filepath.Join(dir, "versions")
	vDir := filepath.Join(versionsDir, "1.5.0")
	if err := os.MkdirAll(vDir, 0o755); err != nil {
		t.Fatal(err)
	}
	createTestELF(t, filepath.Join(vDir, "terraform"), elf.EM_X86_64)

	t.Setenv("TFENV_CONFIG_DIR", dir)
	t.Setenv("TFENV_TERRAFORM_VERSION", "")
	t.Setenv("TFENV_DIR", dir)
	t.Setenv("HOME", dir)

	var code int
	stdout := captureStdout(func() {
		code = RunList(nil)
	})
	if code != 0 {
		t.Fatalf("RunList() returned %d, want 0", code)
	}
	if !strings.Contains(stdout, "  1.5.0 (amd64)") {
		t.Errorf("stdout = %q, want to contain '  1.5.0 (amd64)'", stdout)
	}
	if strings.Contains(stdout, "*") {
		t.Errorf("stdout = %q, should not contain '*' (no default set)", stdout)
	}
}

func TestRunList_UsageError(t *testing.T) {
	code := RunList([]string{"extra"})
	if code != 1 {
		t.Errorf("RunList(extra) returned %d, want 1", code)
	}
}

// --- RunListRemote tests ---

func TestRunListRemote_AllVersions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>
			<a href="/terraform/1.6.0/">terraform_1.6.0</a>
			<a href="/terraform/1.5.0/">terraform_1.5.0</a>
			<a href="/terraform/1.4.0/">terraform_1.4.0</a>
		</body></html>`))
	}))
	t.Cleanup(server.Close)

	dir := t.TempDir()
	t.Setenv("TFENV_CONFIG_DIR", dir)
	t.Setenv("TFENV_REMOTE", server.URL)

	var code int
	stdout := captureStdout(func() {
		code = RunListRemote(nil)
	})
	if code != 0 {
		t.Fatalf("RunListRemote() returned %d, want 0", code)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d: %q", len(lines), stdout)
	}
	if lines[0] != "1.6.0" {
		t.Errorf("line 0 = %q, want %q", lines[0], "1.6.0")
	}
	if lines[1] != "1.5.0" {
		t.Errorf("line 1 = %q, want %q", lines[1], "1.5.0")
	}
	if lines[2] != "1.4.0" {
		t.Errorf("line 2 = %q, want %q", lines[2], "1.4.0")
	}
}

func TestRunListRemote_WithRegex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>
			<a href="/terraform/1.6.0/">terraform_1.6.0</a>
			<a href="/terraform/1.5.3/">terraform_1.5.3</a>
			<a href="/terraform/1.5.0/">terraform_1.5.0</a>
			<a href="/terraform/1.4.0/">terraform_1.4.0</a>
		</body></html>`))
	}))
	t.Cleanup(server.Close)

	dir := t.TempDir()
	t.Setenv("TFENV_CONFIG_DIR", dir)
	t.Setenv("TFENV_REMOTE", server.URL)

	var code int
	stdout := captureStdout(func() {
		code = RunListRemote([]string{`^1\.5`})
	})
	if code != 0 {
		t.Fatalf("RunListRemote(regex) returned %d, want 0", code)
	}

	lines := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d: %q", len(lines), stdout)
	}
	if lines[0] != "1.5.3" {
		t.Errorf("line 0 = %q, want %q", lines[0], "1.5.3")
	}
	if lines[1] != "1.5.0" {
		t.Errorf("line 1 = %q, want %q", lines[1], "1.5.0")
	}
}

func TestRunListRemote_InvalidRegex(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(`<html><body>
			<a href="/terraform/1.5.0/">terraform_1.5.0</a>
		</body></html>`))
	}))
	t.Cleanup(server.Close)

	dir := t.TempDir()
	t.Setenv("TFENV_CONFIG_DIR", dir)
	t.Setenv("TFENV_REMOTE", server.URL)

	code := RunListRemote([]string{"[invalid"})
	if code != 1 {
		t.Errorf("RunListRemote(invalid regex) returned %d, want 1", code)
	}
}

func TestRunListRemote_UsageError(t *testing.T) {
	code := RunListRemote([]string{"arg1", "arg2"})
	if code != 1 {
		t.Errorf("RunListRemote(too many args) returned %d, want 1", code)
	}
}
