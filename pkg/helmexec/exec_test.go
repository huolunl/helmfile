package helmexec

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/Masterminds/semver/v3"
	"go.uber.org/zap"
)

// Mocking the command-line runner

type mockRunner struct {
	output []byte
	err    error
}

func (mock *mockRunner) ExecuteStdIn(cmd string, args []string, env map[string]string, stdin io.Reader) ([]byte, error) {
	return mock.output, mock.err
}

func (mock *mockRunner) Execute(cmd string, args []string, env map[string]string) ([]byte, error) {
	return mock.output, mock.err
}

func MockExecer(logger *zap.SugaredLogger, kubeContext string) *execer {
	execer := New("helm", logger, kubeContext, &mockRunner{}, &bytes.Buffer{})
	return execer
}

// Test methods

func TestNewHelmExec(t *testing.T) {
	buffer := bytes.NewBufferString("something")
	helm := MockExecer(NewLogger(buffer, "debug"), "dev")
	if helm.kubeContext != "dev" {
		t.Error("helmexec.New() - kubeContext")
	}
	if buffer.String() != "something" {
		t.Error("helmexec.New() - changed buffer")
	}
	if len(helm.extra) != 0 {
		t.Error("helmexec.New() - extra args not empty")
	}
}

func Test_SetExtraArgs(t *testing.T) {
	helm := MockExecer(NewLogger(os.Stdout, "info"), "dev")
	helm.SetExtraArgs()
	if len(helm.extra) != 0 {
		t.Error("helmexec.SetExtraArgs() - passing no arguments should not change extra field")
	}
	helm.SetExtraArgs("foo")
	if !reflect.DeepEqual(helm.extra, []string{"foo"}) {
		t.Error("helmexec.SetExtraArgs() - one extra argument missing")
	}
	helm.SetExtraArgs("alpha", "beta")
	if !reflect.DeepEqual(helm.extra, []string{"alpha", "beta"}) {
		t.Error("helmexec.SetExtraArgs() - two extra arguments missing (overwriting the previous value)")
	}
}

func Test_SetHelmBinary(t *testing.T) {
	helm := MockExecer(NewLogger(os.Stdout, "info"), "dev")
	if helm.helmBinary != "helm" {
		t.Error("helmexec.command - default command is not helm")
	}
	helm.SetHelmBinary("foo")
	if helm.helmBinary != "foo" {
		t.Errorf("helmexec.SetHelmBinary() - actual = %s expect = foo", helm.helmBinary)
	}
}

func Test_AddRepo_Helm_3_3_2(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := &execer{
		helmBinary:  "helm",
		version:     *semver.MustParse("3.3.2"),
		logger:      logger,
		kubeContext: "dev",
		runner:      &mockRunner{},
	}
	err := helm.AddRepo("myRepo", "https://repo.example.com/", "", "cert.pem", "key.pem", "", "", "", "", "")
	expected := `Adding repo myRepo https://repo.example.com/
exec: helm --kube-context dev repo add myRepo https://repo.example.com/ --force-update --cert-file cert.pem --key-file key.pem
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if buffer.String() != expected {
		t.Errorf("helmexec.AddRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_AddRepo(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.AddRepo("myRepo", "https://repo.example.com/", "", "cert.pem", "key.pem", "", "", "", "", "")
	expected := `Adding repo myRepo https://repo.example.com/
exec: helm --kube-context dev repo add myRepo https://repo.example.com/ --cert-file cert.pem --key-file key.pem
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if buffer.String() != expected {
		t.Errorf("helmexec.AddRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	err = helm.AddRepo("myRepo", "https://repo.example.com/", "ca.crt", "", "", "", "", "", "", "")
	expected = `Adding repo myRepo https://repo.example.com/
exec: helm --kube-context dev repo add myRepo https://repo.example.com/ --ca-file ca.crt
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if buffer.String() != expected {
		t.Errorf("helmexec.AddRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	err = helm.AddRepo("myRepo", "https://repo.example.com/", "", "", "", "", "", "", "", "")
	expected = `Adding repo myRepo https://repo.example.com/
exec: helm --kube-context dev repo add myRepo https://repo.example.com/
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if buffer.String() != expected {
		t.Errorf("helmexec.AddRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	err = helm.AddRepo("acrRepo", "", "", "", "", "", "", "acr", "", "")
	expected = `Adding repo acrRepo (acr)
exec: az acr helm repo add --name acrRepo
exec: az acr helm repo add --name acrRepo: 
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if buffer.String() != expected {
		t.Errorf("helmexec.AddRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	err = helm.AddRepo("otherRepo", "", "", "", "", "", "", "unknown", "", "")
	expected = `ERROR: unknown type 'unknown' for repository otherRepo
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.AddRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	err = helm.AddRepo("myRepo", "https://repo.example.com/", "", "", "", "example_user", "example_password", "", "", "")
	expected = `Adding repo myRepo https://repo.example.com/
exec: helm --kube-context dev repo add myRepo https://repo.example.com/ --username example_user --password example_password
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.AddRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	err = helm.AddRepo("", "https://repo.example.com/", "", "", "", "", "", "", "", "")
	expected = `empty field name

`
	if err != nil && err.Error() != "empty field name" {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.AddRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	err = helm.AddRepo("myRepo", "https://repo.example.com/", "", "", "", "example_user", "example_password", "", "true", "")
	expected = `Adding repo myRepo https://repo.example.com/
exec: helm --kube-context dev repo add myRepo https://repo.example.com/ --username example_user --password example_password --pass-credentials
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.AddRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	err = helm.AddRepo("myRepo", "https://repo.example.com/", "", "", "", "", "", "", "", "true")
	expected = `Adding repo myRepo https://repo.example.com/
exec: helm --kube-context dev repo add myRepo https://repo.example.com/ --insecure-skip-tls-verify
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.AddRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_UpdateRepo(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.UpdateRepo()
	expected := `Updating repo
exec: helm --kube-context dev repo update
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.UpdateRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_SyncRelease(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.SyncRelease(HelmContext{}, "release", "chart", "--timeout 10", "--wait", "--wait-for-jobs")
	expected := `Upgrading release=release, chart=chart
exec: helm --kube-context dev upgrade --install --reset-values release chart --timeout 10 --wait --wait-for-jobs
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.SyncRelease()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	err = helm.SyncRelease(HelmContext{}, "release", "chart")
	expected = `Upgrading release=release, chart=chart
exec: helm --kube-context dev upgrade --install --reset-values release chart
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.SyncRelease()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_SyncReleaseTillerless(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.SyncRelease(HelmContext{Tillerless: true, TillerNamespace: "foo"}, "release", "chart",
		"--timeout 10", "--wait", "--wait-for-jobs")
	expected := `Upgrading release=release, chart=chart
exec: helm --kube-context dev tiller run foo -- helm upgrade --install --reset-values release chart --timeout 10 --wait --wait-for-jobs
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.SyncRelease()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_UpdateDeps(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.UpdateDeps("./chart/foo")
	expected := `Updating dependency ./chart/foo
exec: helm --kube-context dev dependency update ./chart/foo
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.UpdateDeps()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	helm.SetExtraArgs("--verify")
	err = helm.UpdateDeps("./chart/foo")
	expected = `Updating dependency ./chart/foo
exec: helm --kube-context dev dependency update ./chart/foo --verify
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.AddRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_BuildDeps(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.BuildDeps("foo", "./chart/foo")
	expected := `Building dependency release=foo, chart=./chart/foo
exec: helm --kube-context dev dependency build ./chart/foo
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.BuildDeps()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	helm.SetExtraArgs("--verify")
	err = helm.BuildDeps("foo", "./chart/foo")
	expected = `Building dependency release=foo, chart=./chart/foo
exec: helm --kube-context dev dependency build ./chart/foo --verify
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.BuildDeps()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_DecryptSecret(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")

	tmpFilePath := "path/to/temp/file"
	helm.writeTempFile = func(content []byte) (string, error) {
		return tmpFilePath, nil
	}

	_, err := helm.DecryptSecret(HelmContext{}, "secretName")
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
		} else {
			t.Errorf("Error: %v", err)
		}
	}
	cwd, err := filepath.Abs(".")
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	// Run again for caching
	_, err = helm.DecryptSecret(HelmContext{}, "secretName")

	expected := fmt.Sprintf(`Preparing to decrypt secret %v/secretName
Decrypting secret %s/secretName
exec: helm --kube-context dev secrets dec %s/secretName
Preparing to decrypt secret %s/secretName
Found secret in cache %s/secretName
`, cwd, cwd, cwd, cwd, cwd)
	if err != nil {
		if _, ok := err.(*os.PathError); ok {
		} else {
			t.Errorf("Error: %v", err)
		}
	}
	if d := cmp.Diff(expected, buffer.String()); d != "" {
		t.Errorf("helmexec.DecryptSecret(): want (-), got (+):\n%s", d)
	}
}

func Test_DecryptSecretWithGotmpl(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")

	tmpFilePath := "path/to/temp/file"
	helm.writeTempFile = func(content []byte) (string, error) {
		return tmpFilePath, nil
	}

	secretName := "secretName.yaml.gotmpl"
	_, decryptErr := helm.DecryptSecret(HelmContext{}, secretName)
	cwd, err := filepath.Abs(".")
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	expected := fmt.Sprintf(`%s/%s.yaml.dec`, cwd, secretName)
	if d := cmp.Diff(expected, decryptErr.(*os.PathError).Path); d != "" {
		t.Errorf("helmexec.DecryptSecret(): want (-), got (+):\n%s", d)
	}
}

func Test_DiffRelease(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.DiffRelease(HelmContext{}, "release", "chart", false, "--timeout 10", "--wait", "--wait-for-jobs")
	expected := `Comparing release=release, chart=chart
exec: helm --kube-context dev diff upgrade --reset-values --allow-unreleased release chart --timeout 10 --wait --wait-for-jobs
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.DiffRelease()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	err = helm.DiffRelease(HelmContext{}, "release", "chart", false)
	expected = `Comparing release=release, chart=chart
exec: helm --kube-context dev diff upgrade --reset-values --allow-unreleased release chart
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.DiffRelease()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_DiffReleaseTillerless(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.DiffRelease(HelmContext{Tillerless: true}, "release", "chart", false, "--timeout 10", "--wait", "--wait-for-jobs")
	expected := `Comparing release=release, chart=chart
exec: helm --kube-context dev tiller run -- helm diff upgrade --reset-values --allow-unreleased release chart --timeout 10 --wait --wait-for-jobs
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.DiffRelease()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_DeleteRelease(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.DeleteRelease(HelmContext{}, "release")
	expected := `Deleting release
exec: helm --kube-context dev delete release
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.DeleteRelease()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}
func Test_DeleteRelease_Flags(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.DeleteRelease(HelmContext{}, "release", "--purge")
	expected := `Deleting release
exec: helm --kube-context dev delete release --purge
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.DeleteRelease()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_TestRelease(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.TestRelease(HelmContext{}, "release")
	expected := `Testing release
exec: helm --kube-context dev test release
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.TestRelease()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}
func Test_TestRelease_Flags(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.TestRelease(HelmContext{}, "release", "--cleanup", "--timeout", "60")
	expected := `Testing release
exec: helm --kube-context dev test release --cleanup --timeout 60
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.TestRelease()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_ReleaseStatus(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.ReleaseStatus(HelmContext{}, "myRelease")
	expected := `Getting status myRelease
exec: helm --kube-context dev status myRelease
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.ReleaseStatus()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_exec(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "")
	env := map[string]string{}
	_, err := helm.exec([]string{"version"}, env)
	expected := `exec: helm version
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.exec()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	helm = MockExecer(logger, "dev")
	ret, _ := helm.exec([]string{"diff"}, env)
	if len(ret) != 0 {
		t.Error("helmexec.exec() - expected empty return value")
	}

	buffer.Reset()
	helm = MockExecer(logger, "dev")
	_, err = helm.exec([]string{"diff", "release", "chart", "--timeout 10", "--wait", "--wait-for-jobs"}, env)
	expected = `exec: helm --kube-context dev diff release chart --timeout 10 --wait --wait-for-jobs
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.exec()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	_, err = helm.exec([]string{"version"}, env)
	expected = `exec: helm --kube-context dev version
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.exec()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	helm.SetExtraArgs("foo")
	_, err = helm.exec([]string{"version"}, env)
	expected = `exec: helm --kube-context dev version foo
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.exec()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}

	buffer.Reset()
	helm = MockExecer(logger, "")
	helm.SetHelmBinary("overwritten")
	_, err = helm.exec([]string{"version"}, env)
	expected = `exec: overwritten version
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.exec()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_Lint(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.Lint("release", "path/to/chart", "--values", "file.yml")
	expected := `Linting release=release, chart=path/to/chart
exec: helm --kube-context dev lint path/to/chart --values file.yml
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.Lint()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

func Test_Fetch(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewLogger(&buffer, "debug")
	helm := MockExecer(logger, "dev")
	err := helm.Fetch("chart", "--version", "1.2.3", "--untar", "--untardir", "/tmp/dir")
	expected := `Fetching chart
exec: helm --kube-context dev fetch chart --version 1.2.3 --untar --untardir /tmp/dir
`
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buffer.String() != expected {
		t.Errorf("helmexec.Lint()\nactual = %v\nexpect = %v", buffer.String(), expected)
	}
}

var logLevelTests = map[string]string{
	"debug": `Adding repo myRepo https://repo.example.com/
exec: helm repo add myRepo https://repo.example.com/ --username example_user --password example_password
`,
	"info": `Adding repo myRepo https://repo.example.com/
`,
	"warn": ``,
}

func Test_LogLevels(t *testing.T) {
	var buffer bytes.Buffer
	for logLevel, expected := range logLevelTests {
		buffer.Reset()
		logger := NewLogger(&buffer, logLevel)
		helm := MockExecer(logger, "")
		err := helm.AddRepo("myRepo", "https://repo.example.com/", "", "", "", "example_user", "example_password", "", "", "")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if buffer.String() != expected {
			t.Errorf("helmexec.AddRepo()\nactual = %v\nexpect = %v", buffer.String(), expected)
		}
	}
}
