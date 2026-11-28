package ambient_test

import (
	_ "embed"
	"testing"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsarna/ctyx/envcty"
	_ "github.com/tsarna/vinculum/ambient"
	"github.com/tsarna/vinculum/config"
	"github.com/zclconf/go-cty/cty"
	"go.uber.org/zap"
)

//go:embed testdata/env.vcl
var envtest []byte

func TestEnv(t *testing.T) {
	logger, err := zap.NewDevelopment()
	assert.NoError(t, err)

	_, diags := config.NewConfig().WithSources(envtest).WithLogger(logger).Build()
	if diags.HasErrors() {
		t.Fatalf("failed to build config: %v", diags)
	}
}

func TestEnvInExpressions(t *testing.T) {
	ctx := &hcl.EvalContext{Variables: map[string]cty.Value{
		"env": envcty.MapFrom([]string{"HOME=/home/me", "A.B=dotted", "café=crème"}),
	}}
	eval := func(src string) (cty.Value, hcl.Diagnostics) {
		expr, diags := hclsyntax.ParseExpression([]byte(src), "test", hcl.InitialPos)
		require.False(t, diags.HasErrors(), diags.Error())
		return expr.Value(ctx)
	}

	for src, want := range map[string]string{
		`env.HOME`:    "/home/me",
		`env["A.B"]`:  "dotted",
		`env.café`:    "crème",
		`env["HOME"]`: "/home/me",
	} {
		v, diags := eval(src)
		require.False(t, diags.HasErrors(), "%s: %s", src, diags.Error())
		assert.Equal(t, want, v.AsString(), src)
	}

	_, diags := eval(`env.UNSET`)
	assert.True(t, diags.HasErrors())
}
