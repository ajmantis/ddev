package mssql

import (
	"fmt"
	"time"

	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func pathConfigLease(b *backend) *framework.Path {
	return &framework.Path{
		Pattern: "config/lease",
		Fields: map[string]*framework.FieldSchema{
			"ttl": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Default ttl for roles.",
			},

			"ttl_max": &framework.FieldSchema{
				Type:        framework.TypeString,
				Description: "Maximum time a credential is valid for.",
			},
		},

		Callbacks: map[logical.Operation]framework.OperationFunc{
			logical.ReadOperation:   b.pathConfigLeaseRead,
			logical.UpdateOperation: b.pathConfigLeaseWrite,
		},

		HelpSynopsis:    pathConfigLeaseHelpSyn,
		HelpDescription: pathConfigLeaseHelpDesc,
	}
}

func (b *backend) pathConfigLeaseWrite(
	req *logical.Request, d *framework.FieldData) (*logical.Response, error) {
	ttlRaw := d.Get("ttl").(string)
	ttlMaxRaw := d.Get("ttl_max").(string)

	ttl, err := time.ParseDuration(ttlRaw)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf(
			"Invalid ttl: %s", err)), nil
	}
	ttlMax, err := time.ParseDuration(ttlMaxRaw)
	if err != nil {
		return logical.ErrorResponse(fmt.Sprintf(
			"Invalid ttl_max: %s", err)), nil
	}

	// Store it
	entry, err := logical.StorageEntryJSON("config/lease", &configLease{
		TTL:    ttl,
		TTLMax: ttlMax,
	})
	if err != nil {
		return nil, err
	}
	if err := req.Storage.Put(entry); err != nil {
		return nil, err
	}

	return nil, nil
}

func (b *backend) pathConfigLeaseRead(
	req *logical.Request, data *framework.FieldData) (*logical.Response, error) {
	leaseConfig, err := b.LeaseConfig(req.Storage)

	if err != nil {
		return nil, err
	}
	if leaseConfig == nil {
		return nil, nil
	}

	return &logical.Response{
		Data: map[string]interface{}{
			"ttl":     leaseConfig.TTL.String(),
			"ttl_max": leaseConfig.TTLMax.String(),
		},
	}, nil
}

type configLease struct {
	TTL    time.Duration
	TTLMax time.Duration
}

const pathConfigLeaseHelpSyn = `
Configure the default lease ttl for generated credentials.
`

const pathConfigLeaseHelpDesc = `
This configures the default lease ttl used for credentials
generated by this backend. The ttl specifies the duration that a
credential will be valid for, as well as the maximum session for
a set of credentials.

The format for the ttl is "1h" or integer and then unit. The longest
unit is hour.
`