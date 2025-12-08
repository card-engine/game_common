package message

import (
	pb "github.com/zuodazuoqianggame/game_common/jili/fish/pb"
	"google.golang.org/protobuf/proto"
)

func MsgPack(cmd uint32, data interface{}) ([]byte, error) {
	packet := &pb.Command{
		Type: uint32(cmd),
	}

	if data != nil {
		b, err := proto.Marshal(data.(proto.Message))
		if err != nil {
			return nil, err
		}
		packet.Data = b
	}

	return proto.Marshal(packet)
}

func UnPack(data []byte) (*pb.Command, error) {
	packet := &pb.Command{}
	err := proto.Unmarshal(data, packet)
	return packet, err
}
