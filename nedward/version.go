package nedward

import "github.com/nedscode/nedward/common"

func (c *Client) Version() string {
	return common.NedwardVersion
}
