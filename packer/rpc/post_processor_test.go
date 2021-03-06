package rpc

import (
	"github.com/mitchellh/packer/packer"
	"net/rpc"
	"reflect"
	"testing"
)

var testPostProcessorArtifact = new(testArtifact)

type TestPostProcessor struct {
	configCalled bool
	configVal    []interface{}
	ppCalled     bool
	ppArtifact   packer.Artifact
	ppUi         packer.Ui
}

func (pp *TestPostProcessor) Configure(v ...interface{}) error {
	pp.configCalled = true
	pp.configVal = v
	return nil
}

func (pp *TestPostProcessor) PostProcess(ui packer.Ui, a packer.Artifact) (packer.Artifact, bool, error) {
	pp.ppCalled = true
	pp.ppArtifact = a
	pp.ppUi = ui
	return testPostProcessorArtifact, false, nil
}

func TestPostProcessorRPC(t *testing.T) {
	// Create the interface to test
	p := new(TestPostProcessor)

	// Start the server
	server := rpc.NewServer()
	RegisterPostProcessor(server, p)
	address := serveSingleConn(server)

	// Create the client over RPC and run some methods to verify it works
	client, err := rpc.Dial("tcp", address)
	if err != nil {
		t.Fatalf("Error connecting to rpc: %s", err)
	}

	// Test Configure
	config := 42
	pClient := PostProcessor(client)
	err = pClient.Configure(config)
	if err != nil {
		t.Fatalf("error: %s", err)
	}

	if !p.configCalled {
		t.Fatal("config should be called")
	}

	if !reflect.DeepEqual(p.configVal, []interface{}{42}) {
		t.Fatalf("unknown config value: %#v", p.configVal)
	}

	// Test PostProcess
	a := new(testArtifact)
	ui := new(testUi)
	artifact, _, err := pClient.PostProcess(ui, a)
	if err != nil {
		t.Fatalf("err: %s", err)
	}

	if !p.ppCalled {
		t.Fatal("postprocess should be called")
	}

	if p.ppArtifact.BuilderId() != "bid" {
		t.Fatal("unknown artifact")
	}

	if artifact.BuilderId() != "bid" {
		t.Fatal("unknown result artifact")
	}
}

func TestPostProcessor_Implements(t *testing.T) {
	var raw interface{}
	raw = PostProcessor(nil)
	if _, ok := raw.(packer.PostProcessor); !ok {
		t.Fatal("not a postprocessor")
	}
}
