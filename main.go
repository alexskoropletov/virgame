package main

import (
	"image"
	"log"
	"math/rand"
	"time"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/ebitenutil"
)

type FrameType struct {
	name        string
	imageFrameX float64
	imageFrameY float64
	framesCount int
	isAlive     bool
}

type FieldTile struct {
	xLocation    int
	yLocation    int
	xScreenCoord float64
	yScreenCoord float64
	image        *FrameType
}

type Game struct {
	count       int
	fieldTiles  [][]*FieldTile
	playerTurns int
	enemyTurns  int
}

const (
	fieldScreenCoordX = 10
	fieldScreenCoordY = 10
	fieldWidth        = 5
	fieldHeight       = 6
	tileSize          = 32
	assetTileSize     = 40
)

var (
	err                 error
	tilesImage          *ebiten.Image
	NeutralFrames       = &FrameType{name: "neutral", imageFrameX: 0, imageFrameY: 0, framesCount: 1, isAlive: false}
	PlayerFrames        = &FrameType{name: "player", imageFrameX: 0, imageFrameY: 200, framesCount: 3, isAlive: true}
	EnemyFrames         = &FrameType{name: "enemy", imageFrameX: 0, imageFrameY: 160, framesCount: 3, isAlive: true}
	defaultScreenWidth  = fieldWidth*tileSize + fieldScreenCoordY*2
	defaultScreenHeight = fieldHeight*tileSize + fieldScreenCoordX*2
)

var initialized bool = false

func InitGame(game *Game) *Game {
	tilesImage, _, err = ebitenutil.NewImageFromFile("assets/tiles.png", ebiten.FilterDefault)
	if err != nil {
		log.Fatal(err)
	}

	for x := 0; x < fieldWidth; x++ {
		temp := make([]*FieldTile, 0)
		for y := 0; y < fieldHeight; y++ {
			xCoord := float64(x * tileSize)
			yCoord := float64(y * tileSize)
			temp = append(temp, &FieldTile{xScreenCoord: xCoord, yScreenCoord: yCoord, image: NeutralFrames, xLocation: x, yLocation: y})
		}
		game.fieldTiles = append(game.fieldTiles, temp)
	}
	// -----
	game.fieldTiles[2][0].image = EnemyFrames
	game.fieldTiles[2][fieldHeight-1].image = PlayerFrames
	game.playerTurns = 0
	game.enemyTurns = 2

	return game
}

func GetSubImage(frame *FrameType, frameNum int) *ebiten.Image {
	shift := ((frameNum / 15) % frame.framesCount * assetTileSize)
	x := int(frame.imageFrameX) + shift
	y := int(frame.imageFrameY)
	width := x + assetTileSize
	height := y + assetTileSize
	return tilesImage.SubImage(image.Rect(x, y, width, height)).(*ebiten.Image)
}

func (game *Game) ClickedOnTile(mouseX, mouseY int) bool {
	playerMoveTiles := game.TilesPlayerOrEnemyCanGo("player")
	for _, row := range game.fieldTiles {
		for _, tile := range row {
			topTileX := int(tile.xScreenCoord + fieldScreenCoordX)
			topTileY := int(tile.yScreenCoord + fieldScreenCoordY)
			lowTileX := int(tile.xScreenCoord + fieldScreenCoordX + tileSize)
			lowTileY := int(tile.yScreenCoord + fieldScreenCoordY + tileSize)
			if mouseX >= topTileX && mouseX <= lowTileX {
				if mouseY >= topTileY && mouseY <= lowTileY {
					if CanGo(playerMoveTiles, tile) {
						tile.image = PlayerFrames
						game.playerTurns--
					}
					return true
				}
			}
		}
	}
	return false
}

func CanGo(allowedTiles []*FieldTile, wantedTile *FieldTile) bool {
	for _, item := range allowedTiles {
		if item.xLocation == wantedTile.xLocation && item.yLocation == wantedTile.yLocation {
			return true
		}
	}
	return false
}

func (game *Game) TilesPlayerOrEnemyCanGo(name string) []*FieldTile {
	moveTiles := make([]*FieldTile, 0)
	for x, row := range game.fieldTiles {
		for y, tile := range row {
			if tile.image.name == name {
				if x-1 >= 0 {
					if nextTile := game.fieldTiles[x-1][y]; nextTile != nil && nextTile.image.name == NeutralFrames.name {
						moveTiles = append(moveTiles, nextTile)
					}
				}
				if x+1 < fieldWidth {
					if nextTile := game.fieldTiles[x+1][y]; nextTile != nil && nextTile.image.name == NeutralFrames.name {
						moveTiles = append(moveTiles, nextTile)
					}
				}
				if y-1 >= 0 {
					if nextTile := game.fieldTiles[x][y-1]; nextTile != nil && nextTile.image.name == NeutralFrames.name {
						moveTiles = append(moveTiles, nextTile)
					}
				}
				if y+1 < fieldHeight {
					if nextTile := game.fieldTiles[x][y+1]; nextTile != nil && nextTile.image.name == NeutralFrames.name {
						moveTiles = append(moveTiles, nextTile)
					}
				}
			}
		}
	}
	return moveTiles
}

func (game *Game) DoEnemyMove() {
	enemyMoveTiles := game.TilesPlayerOrEnemyCanGo("enemy")
	if enemyMovesCount := len(enemyMoveTiles); enemyMovesCount > 0 {
		rand.Seed(time.Now().UnixNano())
		enemyMoveTiles[rand.Intn(enemyMovesCount)].image = EnemyFrames
	}
}

func (game *Game) Update(screen *ebiten.Image) error {
	if !initialized {
		game = InitGame(game)
		initialized = true
	}
	game.count++
	return nil
}

func (game *Game) Draw(screen *ebiten.Image) {
	for _, row := range game.fieldTiles {
		for _, tile := range row {
			op := &ebiten.DrawImageOptions{}
			scale := float64(tileSize) / float64(assetTileSize)
			op.GeoM.Scale(scale, scale)
			op.GeoM.Translate(tile.xScreenCoord+fieldScreenCoordX, tile.yScreenCoord+fieldScreenCoordY)
			screen.DrawImage(GetSubImage(tile.image, game.count), op)
		}
	}

	if game.playerTurns > 0 {
		if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			game.ClickedOnTile(ebiten.CursorPosition())
		}
		game.enemyTurns = 2
	} else {
		if game.enemyTurns > 0 {
			if game.count%15 == 1 {
				game.DoEnemyMove()
				game.enemyTurns--
			}
		} else {
			game.playerTurns = 2
		}
	}
}

func (game *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return defaultScreenWidth, defaultScreenHeight
}

func main() {
	ebiten.SetWindowSize(defaultScreenWidth, defaultScreenHeight)
	ebiten.SetWindowTitle("Vir Game")
	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
