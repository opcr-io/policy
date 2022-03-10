package app

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	extendedregistry "github.com/opcr-io/policy/pkg/extended_registry"
	"github.com/pkg/errors"
	"oras.land/oras-go/pkg/content"
)

func (c *PolicyApp) Rm(existingRef string, force bool) error {
	defer c.Cancel()

	existingRefParsed, err := c.calculatePolicyRef(existingRef)
	if err != nil {
		return err
	}

	confirmation := force
	if !force {
		c.UI.Exclamation().
			WithStringValue("reference", existingRefParsed).
			WithAskBoolMap("[Y/n]", &confirmation, map[string]bool{
				"":  true,
				"y": true,
				"n": false,
			}).Msgf("Are you sure?")
	}

	if !confirmation {
		c.UI.Exclamation().Msg("Operation canceled by user.")
		return nil
	}

	ociStore, err := content.NewOCIStore(c.Configuration.PoliciesRoot())
	if err != nil {
		return err
	}
	err = ociStore.LoadIndex()
	if err != nil {
		return err
	}

	existingRefs := ociStore.ListReferences()

	_, ok := existingRefs[existingRefParsed]
	if !ok {
		return errors.Errorf("ref [%s] not found in the local store", existingRef)
	}

	ociStore.DeleteReference(existingRefParsed)

	// TODO: if there are no references left to the policy, perhaps delete the descriptor?
	// or implement a cleanup command

	c.UI.Normal().
		WithStringValue("reference", existingRef).
		Msg("Removed reference.")

	err = ociStore.SaveIndex()
	if err != nil {
		return err
	}

	return nil
}

func (c *PolicyApp) RmRemote(existingRef string, removeAll, force bool) error {
	defer c.Cancel()

	ref, err := c.calculatePolicyRef(existingRef)
	if err != nil {
		return err
	}

	refParsed, err := name.ParseReference(ref)
	if err != nil {
		return errors.Wrapf(err, "invalid reference [%s]", ref)
	}

	server := refParsed.Context().Registry
	creds := c.Configuration.Servers[server.Name()]

	tagsToRemove := []string{}

	confirmation := force
	if !force {
		c.UI.Exclamation().
			WithStringValue("reference", ref).
			WithAskBoolMap("[Y/n]", &confirmation, map[string]bool{
				"":  true,
				"y": true,
				"n": false,
			}).Msgf("Are you sure?")
	}

	if !confirmation {
		c.UI.Exclamation().Msg("Operation canceled by user.")
		return nil
	}

	if removeAll {
		tagsToRemove, err = c.imageTags(refParsed.Context().RegistryStr()+"/"+refParsed.Context().RepositoryStr(), creds.Username, creds.Password)
		if err != nil {
			return err
		}
	} else {
		tagsToRemove = append(tagsToRemove, refParsed.Identifier())
	}

	for _, tag := range tagsToRemove {
		refToRemove := refParsed.Context().Tag(tag)
		c.UI.Normal().Compact().
			WithStringValue("ref", refToRemove.String()).
			Msg("Removing tag.")

		err = remote.Delete(refToRemove,
			remote.WithAuth(&authn.Basic{
				Username: creds.Username,
				Password: creds.Password,
			}),
			remote.WithTransport(c.TransportWithTrustedCAs()))
		if err != nil {
			return errors.Wrapf(err, "failed to delete reference [%s]", ref)
		}
	}

	if removeAll {
		xClient, err := extendedregistry.GetExtendedClient(server.Name(),
			c.Logger,
			&extendedregistry.Config{
				Address:  "https://" + server.Name(),
				Username: creds.Username,
				Password: creds.Password,
			},
			c.TransportWithTrustedCAs())
		if err != nil {
			return errors.Wrap(err, "no extended remove supported")
		}

		policyDef := refParsed.Context().RepositoryStr()
		c.UI.Normal().
			WithStringValue("definition", policyDef).
			Msg("Removing policy definition.")
		err = xClient.RemoveImage(policyDef, "")
		if err != nil {
			return err
		}

	}

	c.UI.Normal().Msg("OK.")

	return nil
}
