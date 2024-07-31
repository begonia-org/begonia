package auth

type Header interface {
	Set(key, value string)
	SendHeader(key, value string)
}

// type GrpcHeader struct {
// 	in  metadata.MD
// 	ctx context.Context
// 	out metadata.MD
// }
// type httpHeader struct {
// 	w http.ResponseWriter
// 	r *http.Request
// }
// type GrpcStreamHeader struct {
// 	*GrpcHeader
// 	ss grpc.ServerStream
// }

// var headerPool = &sync.Pool{
// 	New: func() interface{} {
// 		return &GrpcHeader{}
// 	},
// }
// var httpHeaderPool = &sync.Pool{
// 	New: func() interface{} {
// 		return &httpHeader{}
// 	},
// }

// var grpcStreamHeaderPool = &sync.Pool{
// 	New: func() interface{} {
// 		return &GrpcStreamHeader{}
// 	},
// }

// func (g *GrpcHeader) Release() {
// 	g.ctx = nil
// 	g.in = nil
// 	g.out = nil
// 	headerPool.Put(g)
// }
// func (g *GrpcHeader) Set(key, value string) {
// 	g.in.Set(key, value)
// 	md, _ := metadata.FromIncomingContext(g.ctx)
// 	newMd := metadata.Join(md, g.in)
// 	g.ctx = metadata.NewIncomingContext(g.ctx, newMd)
// 	log.Printf("grpc header set key:%v value:%v", key, value)

// }
// func (g *GrpcHeader) SendHeader(key, value string) {
// 	g.out.Append(key, value)
// 	_ = grpc.SendHeader(g.ctx, g.out)
// 	g.ctx = metadata.NewOutgoingContext(g.ctx, g.out)
// }
// func (g *httpHeader) Release() {
// 	g.w = nil
// 	g.r = nil
// 	httpHeaderPool.Put(g)
// }
// func (g *httpHeader) Set(key, value string) {
// 	g.r.Header.Add(key, value)

// }
// func (g *httpHeader) SendHeader(key, value string) {
// 	g.w.Header().Add(key, value)
// }

// func (g *GrpcStreamHeader) Release() {
// 	g.ctx = nil
// 	g.in = nil
// 	g.out = nil
// 	g.ss = nil
// 	grpcStreamHeaderPool.Put(g)

// }
// func (g *GrpcStreamHeader) Set(key, value string) {
// 	g.in.Set(key, value)
// 	newCtx := metadata.NewIncomingContext(g.ctx, g.in)
// 	g.ctx = newCtx
// 	_ = g.ss.SetHeader(g.in)
// }
// func (g *GrpcStreamHeader) SendHeader(key, value string) {
// 	g.out.Append(key, value)
// 	_ = g.ss.SendHeader(g.out)
// 	g.ctx = metadata.NewOutgoingContext(g.ctx, g.out)
// }

// func (g *GrpcClientStreamHeader) Release() {
// 	g.ctx = nil
// 	g.in = nil
// 	g.out = nil
// 	g.cs = nil
// 	grpcClientHeaderPool.Put(g)

// }
// func (g *GrpcClientStreamHeader) Set(key, value string) {
// 	g.out.Set(key, value)
// 	newCtx := metadata.NewOutgoingContext(g.ctx, g.in)
// 	g.ctx = newCtx
// 	// _ = g.cs.(g.in)

// }
// func NewGrpcHeader(in metadata.MD, ctx context.Context, out metadata.MD) *GrpcHeader {
// 	// return &GrpcHeader{in: in, ctx: ctx, out: out}
// 	header := headerPool.Get().(*GrpcHeader)
// 	header.in = in
// 	header.ctx = ctx
// 	header.out = out
// 	return header
// }
// func NewHttpHeader(w http.ResponseWriter, r *http.Request) *httpHeader {
// 	// return &httpHeader{w: w, r: r}
// 	header := httpHeaderPool.Get().(*httpHeader)
// 	header.w = w
// 	header.r = r
// 	return header
// }

// func NewGrpcStreamHeader(in metadata.MD, ctx context.Context, out metadata.MD, ss grpc.ServerStream) *GrpcStreamHeader {
// 	// return &GrpcStreamHeader{&GrpcHeader{in: in, ctx: ctx, out: out}, ss}
// 	header := grpcStreamHeaderPool.Get().(*GrpcStreamHeader)
// 	header.GrpcHeader = NewGrpcHeader(in, ctx, out)
// 	header.ss = ss
// 	return header
// }
// func NewGrpcClientStreamHeader(in metadata.MD, ctx context.Context, out metadata.MD, cs grpc.ClientStream) *GrpcClientStreamHeader {
// 	header := grpcClientHeaderPool.Get().(*GrpcClientStreamHeader)
// 	header.GrpcHeader = NewGrpcHeader(in, ctx, out)
// 	header.cs = cs
// 	return header
// }
