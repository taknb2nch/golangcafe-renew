package oauth

import (
	"net/url"
	"testing"
)

// https://dev.twitter.com/docs/auth/creating-signature
func TestCalcSignature(t *testing.T) {
	g := NewGenerator(
		"xvz1evFS4wEEPTGEFPHBog",
		"kAcSOqF21Fu85e7zjz7ZN2U4ZRhfV3WpwPAoE3Z7kBw",
		"370773112-GmHxMAgYyLbNEtIKZeRNFsMKPR9EyMZeS9weJAEb",
		"LswwdoUaIvS8ltyTt5jkRh4J50vUPVVHtR2YPi5kE")

	g.SetParam(nonceParam, "kYjzVBB8Y0ZFabxSWbWovY3uYSQ2pTgmZeNu2VS4cg")
	g.SetParam(timestampParam, "1318622958")

	values := make(url.Values)
	values.Add("status", "Hello Ladies + Gentlemen, a signed OAuth request!")

	g.SetUrl("POST", "https://api.twitter.com/1/statuses/update.json?include_entities=true", values)

	g.setDefaultParams()
	signature := g.calcSignature()

	if signature != "tnnArxj06cWHq44gCs1OSKk/jLY=" {
		t.Errorf("\nexpected: %s\nactual  : %s\n", "tnnArxj06cWHq44gCs1OSKk/jLY=", signature)
	}
}
