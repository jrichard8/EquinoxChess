package move

type Move = uint32

const (
	fromShift     = 0
	toShift       = 6
	movingShift   = 12
	capturedShift = 16
	promoShift    = 20
	flagsShift    = 24
)

const (
	fromMask   = 0x3F
	pieceMask  = 0x0F
	flagsMask  = 0x07
	NoPiece    = 0x0F
)

const (
	FlagCastling       = 1 << 0
	FlagEnPassant      = 1 << 1
	FlagDoublePawnPush = 1 << 2
)

func NewMove(from, to, moving, captured, promotion int, flags uint32) uint32 {
	return uint32(from) |
		(uint32(to) << toShift) |
		(uint32(moving) << movingShift) |
		(uint32(captured) << capturedShift) |
		(uint32(promotion) << promoShift) |
		((flags & flagsMask) << flagsShift)
}

func From(move uint32) int {
	return int(move & fromMask)
}

func To(move uint32) int {
	return int((move >> toShift) & fromMask)
}

func MovingPiece(move uint32) int {
	return int((move >> movingShift) & pieceMask)
}

func CapturedPiece(move uint32) int {
	return int((move >> capturedShift) & pieceMask)
}

func PromotionPiece(move uint32) int {
	return int((move >> promoShift) & pieceMask)
}

func IsCapture(move uint32) bool {
	return CapturedPiece(move) != NoPiece
}

func IsPromotion(move uint32) bool {
	return PromotionPiece(move) != NoPiece
}

func IsCastling(move uint32) bool {
	return ((move >> flagsShift) & flagsMask & FlagCastling) != 0
}

func IsEnPassant(move uint32) bool {
	return ((move >> flagsShift) & flagsMask & FlagEnPassant) != 0
}

func IsDoublePawnPush(move uint32) bool {
	return ((move >> flagsShift) & flagsMask & FlagDoublePawnPush) != 0
}
