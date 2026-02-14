package x11

// SpriteData holds pixel data in BGRA format, matching the Framebuffer layout.
type SpriteData struct {
	Width, Height int
	Pixels        []byte // BGRA format, 4 bytes per pixel
}

// BlitSprite draws an entire sprite onto the framebuffer at (dstX, dstY).
func (fb *Framebuffer) BlitSprite(s *SpriteData, dstX, dstY int) {
	fb.BlitSpriteRegion(s, dstX, dstY, 0, 0, s.Width, s.Height)
}

// BlitSpriteRegion draws a sub-region of a sprite onto the framebuffer.
// The source region is defined by (srcX, srcY, srcW, srcH) within the sprite.
// It is placed at (dstX, dstY) on the framebuffer. All clipping is done
// up front so the inner loop has zero bounds checks.
func (fb *Framebuffer) BlitSpriteRegion(s *SpriteData, dstX, dstY, srcX, srcY, srcW, srcH int) {
	// Clip source region to sprite bounds
	if srcX < 0 {
		srcW += srcX
		dstX -= srcX
		srcX = 0
	}
	if srcY < 0 {
		srcH += srcY
		dstY -= srcY
		srcY = 0
	}
	if srcX+srcW > s.Width {
		srcW = s.Width - srcX
	}
	if srcY+srcH > s.Height {
		srcH = s.Height - srcY
	}

	// Clip destination against framebuffer edges
	if dstX < 0 {
		srcX -= dstX
		srcW += dstX
		dstX = 0
	}
	if dstY < 0 {
		srcY -= dstY
		srcH += dstY
		dstY = 0
	}
	if dstX+srcW > fb.Width {
		srcW = fb.Width - dstX
	}
	if dstY+srcH > fb.Height {
		srcH = fb.Height - dstY
	}

	// Nothing to draw after clipping
	if srcW <= 0 || srcH <= 0 {
		return
	}

	fbStride := fb.Width * 4
	spStride := s.Width * 4
	fbPix := fb.Pixels
	spPix := s.Pixels

	for row := 0; row < srcH; row++ {
		fbOff := (dstY+row)*fbStride + dstX*4
		spOff := (srcY+row)*spStride + srcX*4

		for col := 0; col < srcW; col++ {
			a := uint32(spPix[spOff+3])

			if a == 0 {
				// Fully transparent — skip
				fbOff += 4
				spOff += 4
				continue
			}

			if a == 255 {
				// Fully opaque — direct copy (B, G, R)
				fbPix[fbOff] = spPix[spOff]
				fbPix[fbOff+1] = spPix[spOff+1]
				fbPix[fbOff+2] = spPix[spOff+2]
				fbOff += 4
				spOff += 4
				continue
			}

			// Alpha blend: out = (src*a + dst*(255-a) + 1 + ((src*a + dst*(255-a)) >> 8)) >> 8
			invA := 255 - a
			for ch := 0; ch < 3; ch++ {
				s := uint32(spPix[spOff+ch])
				d := uint32(fbPix[fbOff+ch])
				v := s*a + d*invA
				fbPix[fbOff+ch] = uint8((v + 1 + (v >> 8)) >> 8)
			}

			fbOff += 4
			spOff += 4
		}
	}
}
