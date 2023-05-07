package loader

import (
	"embed"
	"image"
	"image/color"
	"io"
	"log"
	"path"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/mp3"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

//go:embed resources
var Embedded embed.FS

const (
	//--RaycastEngine constants
	//--set constant, texture size to be the wall (and sprite) texture size--//
	TexWidth = 256

	// distance to keep away from walls and obstacles to avoid clipping
	// TODO: may want a smaller distance to test vs. sprites
	ClipDistance = 0.1
)

// loadContent will be called once per game and is the place to load
// all of your content.
func LoadContent(mapObj *Map) *TextureHandler {

	// TODO: make resource management better
	tex := NewTextureHandler(mapObj, 32)
	// load wall textures
	tex.Textures[0] = GetTextureFromFile("stone.png")
	tex.Textures[1] = GetTextureFromFile("left_bot_house.png")
	tex.Textures[2] = GetTextureFromFile("right_bot_house.png")
	tex.Textures[3] = GetTextureFromFile("left_top_house.png")
	tex.Textures[4] = GetTextureFromFile("right_top_house.png")
	tex.Textures[5] = GetTextureFromFile("ebitengine_splash.png")

	// separating sprites out a bit from wall textures
	tex.Textures[8] = GetSpriteFromFile("large_rock.png")
	tex.Textures[9] = GetSpriteFromFile("tree_09.png")
	tex.Textures[10] = GetSpriteFromFile("tree_10.png")
	tex.Textures[14] = GetSpriteFromFile("tree_14.png")

	// load texture sheets
	tex.Textures[15] = GetSpriteFromFile("sorcerer_sheet.png")
	tex.Textures[16] = GetSpriteFromFile("crosshairs_sheet.png")
	tex.Textures[17] = GetSpriteFromFile("charged_bolt_sheet.png")
	tex.Textures[18] = GetSpriteFromFile("blue_explosion_sheet.png")
	tex.Textures[19] = GetSpriteFromFile("outleader_walking_sheet.png")
	tex.Textures[20] = GetSpriteFromFile("hand_spell.png")
	tex.Textures[21] = GetSpriteFromFile("hand_staff.png")
	tex.Textures[22] = GetSpriteFromFile("red_bolt.png")
	tex.Textures[23] = GetSpriteFromFile("red_explosion_sheet.png")
	tex.Textures[24] = GetSpriteFromFile("bat_sheet.png")

	// just setting the grass texture apart from the rest since it gets special handling
	return tex
}

func NewImageFromFile(path string) (*ebiten.Image, image.Image, error) {
	f, err := Embedded.Open(filepath.ToSlash(path))
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	eb, im, err := ebitenutil.NewImageFromReader(f)
	return eb, im, err
}

func GetRGBAFromFile(texFile string) *image.RGBA {
	var rgba *image.RGBA
	_, tex, err := NewImageFromFile("resources/textures/" + texFile)
	if err != nil {
		log.Fatal(err)
	}
	if tex != nil {
		rgba = image.NewRGBA(image.Rect(0, 0, TexWidth, TexWidth))
		// convert into RGBA format
		for x := 0; x < TexWidth; x++ {
			for y := 0; y < TexWidth; y++ {
				clr := tex.At(x, y).(color.RGBA)
				rgba.SetRGBA(x, y, clr)
			}
		}
	}

	return rgba
}

func GetTextureFromFile(texFile string) *ebiten.Image {
	eImg, _, err := NewImageFromFile("resources/textures/" + texFile)
	if err != nil {
		log.Fatal(err)
	}
	return eImg
}

func GetSpriteFromFile(sFile string) *ebiten.Image {
	eImg, _, err := NewImageFromFile("resources/sprites/" + sFile)
	if err != nil {
		log.Fatal(err)
	}
	return eImg
}

// Todo
// func (g *Game) loadSprites() {
// 	g.projectiles = make(map[*model.Projectile]struct{}, 1024)
// 	g.effects = make(map[*model.Effect]struct{}, 1024)
// 	g.sprites = make(map[*Sprite]struct{}, 128)

// 	// colors for minimap representation
// 	blueish := color.RGBA{62, 62, 100, 96}
// 	reddish := color.RGBA{180, 62, 62, 96}
// 	brown := color.RGBA{47, 40, 30, 196}
// 	green := color.RGBA{27, 37, 7, 196}
// 	orange := color.RGBA{69, 30, 5, 196}
// 	yellow := color.RGBA{255, 200, 0, 196}

// 	// preload projectile sprites
// 	chargedBoltImg := g.tex.textures[17]
// 	chargedBoltWidth := chargedBoltImg.Bounds().Dx()
// 	chargedBoltCols, chargedBoltRows := 6, 1
// 	chargedBoltScale := 0.3
// 	// in pixels, radius to use for collision testing
// 	chargedBoltPxRadius := 50.0
// 	chargedBoltCollisionRadius := (chargedBoltScale * chargedBoltPxRadius) / (float64(chargedBoltWidth) / float64(chargedBoltCols))
// 	chargedBoltCollisionHeight := 2 * chargedBoltCollisionRadius
// 	chargedBoltProjectile := model.NewAnimatedProjectile(
// 		0, 0, chargedBoltScale, 1, chargedBoltImg, blueish,
// 		chargedBoltCols, chargedBoltRows, raycaster.AnchorCenter, chargedBoltCollisionRadius, chargedBoltCollisionHeight,
// 	)

// 	redBoltImg := g.tex.textures[22]
// 	redBoltWidth := redBoltImg.Bounds().Dx()
// 	redBoltScale := 0.25
// 	// in pixels, radius to use for collision testing
// 	redBoltPxRadius := 4.0
// 	redBoltCollisionRadius := (redBoltScale * redBoltPxRadius) / float64(redBoltWidth)
// 	redBoltCollisionHeight := 2 * redBoltCollisionRadius
// 	redBoltProjectile := model.ProjectileTemplate(
// 		0, 0, redBoltScale, redBoltImg, reddish,
// 		raycaster.AnchorCenter, redBoltCollisionRadius, redBoltCollisionHeight,
// 	)

// 	// preload effect sprites
// 	blueExplosionEffect := model.NewAnimatedEffect(
// 		0, 0, 0.75, 3, g.tex.textures[18], 5, 3, raycaster.AnchorCenter, 1,
// 	)
// 	chargedBoltProjectile.ImpactEffect = *blueExplosionEffect

// 	redExplosionEffect := model.NewAnimatedEffect(
// 		0, 0, 0.20, 1, g.tex.textures[23], 8, 3, raycaster.AnchorCenter, 1,
// 	)
// 	redBoltProjectile.ImpactEffect = *redExplosionEffect

// 	// create weapons
// 	chargedBoltRoF := 2.5      // Rate of Fire (as RoF/second)
// 	chargedBoltVelocity := 6.0 // Velocity (as distance travelled/second)
// 	chargedBoltWeapon := model.NewAnimatedWeapon(1, 1, 1.0, 7, g.tex.textures[20], 3, 1, *chargedBoltProjectile, chargedBoltVelocity, chargedBoltRoF)
// 	g.player.AddWeapon(chargedBoltWeapon)

// 	staffBoltRoF := 6.0
// 	staffBoltVelocity := 24.0
// 	staffBoltWeapon := model.NewAnimatedWeapon(1, 1, 1.0, 7, g.tex.textures[21], 3, 1, *redBoltProjectile, staffBoltVelocity, staffBoltRoF)
// 	g.player.AddWeapon(staffBoltWeapon)

// 	// animated single facing sorcerer
// 	sorcImg := g.tex.textures[15]
// 	sorcWidth, sorcHeight := sorcImg.Bounds().Dx(), sorcImg.Bounds().Dy()
// 	sorcCols, sorcRows := 10, 1
// 	sorcScale := 1.25
// 	// in pixels, radius and height to use for collision testing
// 	sorcPxRadius, sorcPxHeight := 40.0, 120.0
// 	// convert pixel to grid using image pixel size
// 	sorcCollisionRadius := (sorcScale * sorcPxRadius) / (float64(sorcWidth) / float64(sorcCols))
// 	sorcCollisionHeight := (sorcScale * sorcPxHeight) / (float64(sorcHeight) / float64(sorcRows))
// 	sorc := model.NewAnimatedSprite(
// 		22.5, 11.75, sorcScale, 5, sorcImg, yellow, sorcCols, sorcRows, raycaster.AnchorBottom, sorcCollisionRadius, sorcCollisionHeight,
// 	)
// 	// give sprite a sample velocity for movement
// 	sorc.Angle = geom.Radians(180)
// 	sorc.Velocity = 0.02
// 	g.addSprite(sorc)

// 	// animated walking 8-directional sprite character
// 	// [walkerTexFacingMap] player facing angle : texture row index
// 	var walkerTexFacingMap = map[float64]int{
// 		geom.Radians(315): 0,
// 		geom.Radians(270): 1,
// 		geom.Radians(225): 2,
// 		geom.Radians(180): 3,
// 		geom.Radians(135): 4,
// 		geom.Radians(90):  5,
// 		geom.Radians(45):  6,
// 		geom.Radians(0):   7,
// 	}
// 	walkerImg := g.tex.textures[19]
// 	walkerWidth, walkerHeight := walkerImg.Bounds().Dx(), walkerImg.Bounds().Dy()
// 	walkerCols, walkerRows := 4, 8
// 	walkerScale := 0.75
// 	// in pixels, radius and height to use for collision testing
// 	walkerPxRadius, walkerPxHeight := 30.0, 80.0
// 	// convert pixel to grid using image pixel size
// 	walkerCollisionRadius := (walkerScale * walkerPxRadius) / (float64(walkerWidth) / float64(walkerCols))
// 	walkerCollisionHeight := (walkerScale * walkerPxHeight) / (float64(walkerHeight) / float64(walkerRows))
// 	walker := model.NewAnimatedSprite(
// 		7.5, 6.0, walkerScale, 10, walkerImg, yellow, walkerCols, walkerRows, raycaster.AnchorBottom, walkerCollisionRadius, walkerCollisionHeight,
// 	)
// 	walker.SetAnimationReversed(true) // this sprite sheet has reversed animation frame order
// 	walker.SetTextureFacingMap(walkerTexFacingMap)
// 	// give sprite a sample velocity for movement
// 	walker.Angle = geom.Radians(0)
// 	walker.Velocity = 0.02
// 	g.addSprite(walker)

// 	// animated flying 4-directional sprite creature
// 	// [batTexFacingMap] player facing angle : texture row index
// 	var batTexFacingMap = map[float64]int{
// 		geom.Radians(270): 1,
// 		geom.Radians(180): 2,
// 		geom.Radians(90):  3,
// 		geom.Radians(0):   0,
// 	}
// 	batImg := g.tex.textures[24]
// 	batWidth, batHeight := batImg.Bounds().Dx(), batImg.Bounds().Dy()
// 	batCols, batRows := 3, 4
// 	batScale := 0.25
// 	// in pixels, radius and height to use for collision testing
// 	batPxRadius, batPxHeight := 14.0, 25.0
// 	// convert pixel to grid using image pixel size
// 	batCollisionRadius := (batScale * batPxRadius) / (float64(batWidth) / float64(batCols))
// 	batCollisionHeight := (batScale * batPxHeight) / (float64(batHeight) / float64(batRows))
// 	batty := model.NewAnimatedSprite(
// 		10.0, 5.0, batScale, 10, batImg, yellow, batCols, batRows, raycaster.AnchorTop, batCollisionRadius, batCollisionHeight,
// 	)
// 	batty.SetTextureFacingMap(batTexFacingMap)
// 	// raising Z-position of sprite model but using raycaster.AnchorTop to show below that position
// 	batty.Position.Z = 1.0
// 	// give sprite a sample velocity for movement
// 	batty.Angle = geom.Radians(150)
// 	batty.Velocity = 0.03
// 	g.addSprite(batty)

// 	if g.debug {
// 		// just some debugging stuff
// 		sorc.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
// 		walker.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
// 		batty.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
// 		chargedBoltProjectile.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
// 		redBoltProjectile.AddDebugLines(2, color.RGBA{0, 255, 0, 255})
// 	}

// 	// rock that can be jumped over but not walked through
// 	rockImg := g.tex.textures[8]
// 	rockWidth, rockHeight := rockImg.Bounds().Dx(), rockImg.Bounds().Dy()
// 	rockScale := 0.4
// 	rockPxRadius, rockPxHeight := 24.0, 35.0
// 	rockCollisionRadius := (rockScale * rockPxRadius) / float64(rockWidth)
// 	rockCollisionHeight := (rockScale * rockPxHeight) / float64(rockHeight)
// 	rock := model.NewSprite(8.0, 5.5, rockScale, rockImg, brown, raycaster.AnchorBottom, rockCollisionRadius, rockCollisionHeight)
// 	g.addSprite(rock)

// 	// testing sprite scaling
// 	testScale := 0.5
// 	g.addSprite(model.NewSprite(10.5, 2.5, testScale, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))

// 	// // line of trees for testing in front of initial view
// 	// Setting CollisionRadius=0 to disable collision against small trees
// 	g.addSprite(model.NewSprite(19.5, 11.5, 1.0, g.tex.textures[10], brown, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(17.5, 11.5, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(15.5, 11.5, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	// // // render a forest!
// 	g.addSprite(model.NewSprite(11.5, 1.5, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(12.5, 1.5, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(132.5, 1.5, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(11.5, 2, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(12.5, 2, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.5, 2, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(11.5, 2.5, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(12.25, 2.5, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.5, 2.25, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(11.5, 3, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(12.5, 3, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.25, 3, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(10.5, 3.5, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(11.5, 3.25, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(12.5, 3.5, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.25, 3.5, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(10.5, 4, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(11.5, 4, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(12.5, 4, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.5, 4, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(10.5, 4.5, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(11.25, 4.5, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(12.5, 4.5, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.5, 4.5, 1.0, g.tex.textures[10], brown, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(14.5, 4.25, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(10.5, 5, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(11.5, 5, 1.0, g.tex.textures[9], green, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(12.5, 5, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.25, 5, 1.0, g.tex.textures[10], brown, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(14.5, 5, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(11.5, 5.5, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(12.5, 5.25, 1.0, g.tex.textures[10], brown, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.5, 5.25, 1.0, g.tex.textures[10], brown, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(14.5, 5.5, 1.0, g.tex.textures[10], brown, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(15.5, 5.5, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(11.5, 6, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(12.5, 6, 1.0, g.tex.textures[10], brown, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.25, 6, 1.0, g.tex.textures[10], brown, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(14.25, 6, 1.0, g.tex.textures[10], brown, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(15.5, 6, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(12.5, 6.5, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.5, 6.25, 1.0, g.tex.textures[10], brown, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(14.5, 6.5, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(12.5, 7, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.5, 7, 1.0, g.tex.textures[10], brown, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(14.5, 7, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.5, 7.5, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// 	g.addSprite(model.NewSprite(13.5, 8, 1.0, g.tex.textures[14], orange, raycaster.AnchorBottom, 0, 0))
// }

// func (g *Game) addSprite(sprite *Sprite) {
// 	g.sprites[sprite] = struct{}{}
// }

// // func (g *Game) deleteSprite(sprite *Sprite) {
// // 	delete(g.sprites, sprite)
// // }

// func (g *Game) addEffect(effect *model.Effect) {
// 	g.effects[effect] = struct{}{}
// }

// func (g *Game) deleteEffect(effect *model.Effect) {
// 	delete(g.effects, effect)
// }

const (
	audioContextSampleRate = 48000
)

var audioContext = audio.NewContext(audioContextSampleRate)

func LoadAudioPlayer(fname string) *audio.Player {
	if len(fname) == 0 {
		return nil
	}

	f, err := Embedded.Open("resources/audio/" + fname)
	if err != nil {
		log.Fatalf("Load audio file fail: %e", err)
	}

	// In this example, embedded resource "Jab_wav" is used.
	//
	// If you want to use a wav file, open this and pass the file stream to wav.Decode.
	// Note that file's Close() should not be closed here
	// since audio.Player manages stream state.
	//
	//     f, err := os.Open("jab.wav")
	//     if err != nil {
	//         return err
	//     }
	//
	//     d, err := wav.DecodeWithoutResampling(f)
	//     ...

	// Decode wav-formatted data and retrieve decoded PCM stream.

	var d io.Reader
	switch ext := path.Ext(fname); ext {
	case ".wav":
		d, err = wav.DecodeWithSampleRate(audioContextSampleRate, f)
	case ".mp3":
		d, err = mp3.DecodeWithSampleRate(audioContextSampleRate, f)
	default:
		log.Fatal("Unsupported audo file ext ", ext)
	}
	if err != nil {
		log.Fatalf("Decode audio file fail: %e", err)
	}
	// Create an audio.Player that has one stream.
	audioPlayer, err := audioContext.NewPlayer(d)
	if err != nil {
		log.Fatalf("Create audio player fail: %e", err)
	}
	return audioPlayer
}
