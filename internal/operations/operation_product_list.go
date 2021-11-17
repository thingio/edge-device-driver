package operations

import "errors"

// ListProducts for the device manager puts products into the message bus.
func (c *deviceManagerProductOperationClient) ListProducts(protocolID string) error {
	return errors.New("implement me")
}

// ListProducts for the device service takes products from the message bus.
func (c *deviceServiceProductOperationClient) ListProducts(protocolID string) error {
	return errors.New("implement me")
}
