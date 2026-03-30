package ws

// Room 数据平面的代表
// @fake 整个 Room 逻辑和生命周期将会在阶段五进一步开发，在此仅占位供 Hub 转发请求以确保阶段四核心能连通
type Room struct {
	register chan *Client
	leave    chan *Client
	Inbound  chan *Envelope
	clients  map[*Client]bool
}

// NewRoom 初始化
func NewRoom(roomId int, hub *Hub) *Room {
	return &Room{
		register: make(chan *Client),
		leave:    make(chan *Client),
		Inbound:  make(chan *Envelope, 256),
		clients:  make(map[*Client]bool),
	}
}

// Run 启动房间的主事件循环
func (r *Room) Run() {
	for {
		select {
		case c := <-r.register:
			r.clients[c] = true
		case c := <-r.leave:
			if _, ok := r.clients[c]; ok {
				delete(r.clients, c)
			}
		case <-r.Inbound:
			// discard fake
		}
	}
}
