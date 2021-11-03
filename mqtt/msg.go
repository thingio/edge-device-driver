package mqtt

import "fmt"

type Msg struct {
	Topic   string
	Payload []byte
}

func (this *Msg) String() string {
	return fmt.Sprintf("[%s] %s", this.Topic, this.Payload)
}
