package main
import(
	"context"
	fmt "fmt"
	
	pb "route_chat/route_chat"
	"sync"
	"google.golang.org/grpc"
	"crypto/sha256"
	"encoding/hex"
	"bufio"
	"os"
	"time"
	"flag"
	
)


func connect(user *pb.User,client pb.BroadcastClient) error{
	var streamError error
	done:=make(chan int)
	wait:=sync.WaitGroup{}
	fmt.Println(user)
	stream,err := client.CreateStream(context.Background(),&pb.Connect{
		User: user,
		Active:true,
	})
	if err!=nil{
		fmt.Printf("Connect to server fail : %v",err)
	}
	wait.Add(1)
	// receive stream 
	go func (str pb.Broadcast_CreateStreamClient){
		wait.Done()
		for {
			msg,err :=str.Recv()
			if err !=nil{
				fmt.Printf("error reading message %v",err)
				streamError=fmt.Errorf("Error reading message :%v ",err)
				break
			}
			fmt.Printf("%v : %s \n ",msg.User.DisplayName,msg.Message)
		}
	}(stream)
	
	wait.Add(1)
	go func(){
		name :=flag.String("x","VinhDOngdo","")
		defer wait.Done()
		scanner:=bufio.NewScanner(os.Stdin)
		ts :=time.Now()
		msgID:=sha256.Sum256([]byte(ts.String()+ *name))
		for scanner.Scan(){
			msg:=&pb.Message{
				Id: hex.EncodeToString(msgID[:]),
				User:user,
				Message: scanner.Text(),
				Timestamp: ts.String(), 

			}
			_,err:=client.BroadcastMessage(context.Background(),msg)
			if err !=nil{
				fmt.Printf("Error sending message %v ",err)
				
				break
			}
		}
	}()
	go func() {
		wait.Wait()
		close(done)
	 }()
	 
	 <-done
	return streamError

}

func main(){
	name :=flag.String("N","VinhDOngdo","")
	ts:=time.Now()
	id := sha256.Sum256([]byte(ts.String()+ *name))
	user:=&pb.User{
		Id: hex.EncodeToString(id[:]),
		DisplayName: *name,
	}

	conn,err:=grpc.Dial("localhost:8080",grpc.WithInsecure())
	if err != nil {
		fmt.Print("could not connect to server at localhost 8080")

	}
	client:=pb.NewBroadcastClient(conn)
	connect(user, client)

}