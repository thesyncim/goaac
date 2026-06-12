package fdkaac

func FFT4(x []FixpDBL) {
	var a00, a10, a20, a30, tmp0, tmp1 FixpDBL

	a00 = (x[0] + x[4]) >> 1
	a10 = (x[2] + x[6]) >> 1
	a20 = (x[1] + x[5]) >> 1
	a30 = (x[3] + x[7]) >> 1

	x[0] = a00 + a10
	x[1] = a20 + a30

	tmp0 = a00 - x[4]
	tmp1 = a20 - x[5]

	x[4] = a00 - a10
	x[5] = a20 - a30

	a10 = a10 - x[6]
	a30 = a30 - x[7]

	x[2] = tmp0 + a30
	x[6] = tmp0 - a30
	x[3] = tmp1 - a10
	x[7] = tmp1 + a10
}

func FFT8(x []FixpDBL) {
	wPiFourth := FixpSPK{Re: STC(0x5a82799a), Im: STC(0x5a82799a)}

	var a00, a10, a20, a30 FixpDBL
	var y [16]FixpDBL

	a00 = (x[0] + x[8]) >> 1
	a10 = x[4] + x[12]
	a20 = (x[1] + x[9]) >> 1
	a30 = x[5] + x[13]

	y[0] = a00 + (a10 >> 1)
	y[4] = a00 - (a10 >> 1)
	y[1] = a20 + (a30 >> 1)
	y[5] = a20 - (a30 >> 1)

	a00 = a00 - x[8]
	a10 = (a10 >> 1) - x[12]
	a20 = a20 - x[9]
	a30 = (a30 >> 1) - x[13]

	y[2] = a00 + a30
	y[6] = a00 - a30
	y[3] = a20 - a10
	y[7] = a20 + a10

	a00 = (x[2] + x[10]) >> 1
	a10 = x[6] + x[14]
	a20 = (x[3] + x[11]) >> 1
	a30 = x[7] + x[15]

	y[8] = a00 + (a10 >> 1)
	y[12] = a00 - (a10 >> 1)
	y[9] = a20 + (a30 >> 1)
	y[13] = a20 - (a30 >> 1)

	a00 = a00 - x[10]
	a10 = (a10 >> 1) - x[14]
	a20 = a20 - x[11]
	a30 = (a30 >> 1) - x[15]

	y[10] = a00 + a30
	y[14] = a00 - a30
	y[11] = a20 - a10
	y[15] = a20 + a10

	var vr, vi, ur, ui FixpDBL

	ur = y[0] >> 1
	ui = y[1] >> 1
	vr = y[8]
	vi = y[9]
	x[0] = ur + (vr >> 1)
	x[1] = ui + (vi >> 1)
	x[8] = ur - (vr >> 1)
	x[9] = ui - (vi >> 1)

	ur = y[4] >> 1
	ui = y[5] >> 1
	vi = y[12]
	vr = y[13]
	x[4] = ur + (vr >> 1)
	x[5] = ui - (vi >> 1)
	x[12] = ur - (vr >> 1)
	x[13] = ui + (vi >> 1)

	ur = y[10]
	ui = y[11]
	vi, vr = CplxMultDiv2SPK(ui, ur, wPiFourth)

	ur = y[2]
	ui = y[3]
	x[2] = (ur >> 1) + vr
	x[3] = (ui >> 1) + vi
	x[10] = (ur >> 1) - vr
	x[11] = (ui >> 1) - vi

	ur = y[14]
	ui = y[15]
	vr, vi = CplxMultDiv2SPK(ui, ur, wPiFourth)

	ur = y[6]
	ui = y[7]
	x[6] = (ur >> 1) + vr
	x[7] = (ui >> 1) - vi
	x[14] = (ur >> 1) - vr
	x[15] = (ui >> 1) + vi
}

func Scramble(x []FixpDBL, length int) {
	var m, k, j int
	for m = 1; m < length-1; m++ {
		for k = length >> 1; !(((j ^ k) & k) != 0); k >>= 1 {
			j ^= k
		}
		j ^= k

		if j > m {
			tmp := x[2*m]
			x[2*m] = x[2*j]
			x[2*j] = tmp

			tmp = x[2*m+1]
			x[2*m+1] = x[2*j+1]
			x[2*j+1] = tmp
		}
	}
}

func DITFFT(x []FixpDBL, ldn int, trigdata []FixpSPK, trigDataSize int) {
	if ldn < 3 {
		panic("fdkaac: DITFFT requires ldn >= 3")
	}

	n := 1 << uint(ldn)
	Scramble(x, n)

	for i := 0; i < n*2; i += 8 {
		var a00, a10, a20, a30 FixpDBL
		a00 = (x[i+0] + x[i+2]) >> 1
		a10 = (x[i+4] + x[i+6]) >> 1
		a20 = (x[i+1] + x[i+3]) >> 1
		a30 = (x[i+5] + x[i+7]) >> 1

		x[i+0] = a00 + a10
		x[i+4] = a00 - a10
		x[i+1] = a20 + a30
		x[i+5] = a20 - a30

		a00 = a00 - x[i+2]
		a10 = a10 - x[i+6]
		a20 = a20 - x[i+3]
		a30 = a30 - x[i+7]

		x[i+2] = a00 + a30
		x[i+6] = a00 - a30
		x[i+3] = a20 - a10
		x[i+7] = a20 + a10
	}

	mh := 1 << 1
	ldm := ldn - 2
	trigstep := trigDataSize

	for {
		pTrigData := 0
		mh <<= 1
		trigstep >>= 1

		{
			xt1 := 0
			r := n
			for {
				xt2 := xt1 + (mh << 1)
				var vr, vi, ur, ui FixpDBL

				vi = x[xt2+1] >> 1
				vr = x[xt2+0] >> 1

				ur = x[xt1+0] >> 1
				ui = x[xt1+1] >> 1

				x[xt1+0] = ur + vr
				x[xt1+1] = ui + vi

				x[xt2+0] = ur - vr
				x[xt2+1] = ui - vi

				xt1 += mh
				xt2 += mh

				vr = x[xt2+1] >> 1
				vi = x[xt2+0] >> 1

				ur = x[xt1+0] >> 1
				ui = x[xt1+1] >> 1

				x[xt1+0] = ur + vr
				x[xt1+1] = ui - vi

				x[xt2+0] = ur - vr
				x[xt2+1] = ui + vi

				xt1 = xt2 + mh
				r -= mh << 1
				if r == 0 {
					break
				}
			}
		}

		for j := 4; j < mh; j += 4 {
			xt1 := j >> 1
			pTrigData += trigstep
			cs := trigdata[pTrigData]
			r := n

			for {
				xt2 := xt1 + (mh << 1)
				var vr, vi, ur, ui FixpDBL

				vi, vr = CplxMultDiv2SPK(x[xt2+1], x[xt2+0], cs)

				ur = x[xt1+0] >> 1
				ui = x[xt1+1] >> 1

				x[xt1+0] = ur + vr
				x[xt1+1] = ui + vi

				x[xt2+0] = ur - vr
				x[xt2+1] = ui - vi

				xt1 += mh
				xt2 += mh

				vr, vi = CplxMultDiv2SPK(x[xt2+1], x[xt2+0], cs)

				ur = x[xt1+0] >> 1
				ui = x[xt1+1] >> 1

				x[xt1+0] = ur + vr
				x[xt1+1] = ui - vi

				x[xt2+0] = ur - vr
				x[xt2+1] = ui + vi

				xt1 -= j
				xt2 = xt1 + (mh << 1)

				vi, vr = CplxMultDiv2SPK(x[xt2+0], x[xt2+1], cs)

				ur = x[xt1+0] >> 1
				ui = x[xt1+1] >> 1

				x[xt1+0] = ur + vr
				x[xt1+1] = ui - vi

				x[xt2+0] = ur - vr
				x[xt2+1] = ui + vi

				xt1 += mh
				xt2 += mh

				vr, vi = CplxMultDiv2SPK(x[xt2+0], x[xt2+1], cs)

				ur = x[xt1+0] >> 1
				ui = x[xt1+1] >> 1

				x[xt1+0] = ur - vr
				x[xt1+1] = ui - vi

				x[xt2+0] = ur + vr
				x[xt2+1] = ui + vi

				xt1 = xt2 + j
				r -= mh << 1
				if r == 0 {
					break
				}
			}
		}

		{
			xt1 := mh >> 1
			r := n
			wPiFourth := FixpSPK{Re: STC(0x5a82799a), Im: STC(0x5a82799a)}

			for {
				xt2 := xt1 + (mh << 1)
				var vr, vi, ur, ui FixpDBL

				vi, vr = CplxMultDiv2SPK(x[xt2+1], x[xt2+0], wPiFourth)

				ur = x[xt1+0] >> 1
				ui = x[xt1+1] >> 1

				x[xt1+0] = ur + vr
				x[xt1+1] = ui + vi

				x[xt2+0] = ur - vr
				x[xt2+1] = ui - vi

				xt1 += mh
				xt2 += mh

				vr, vi = CplxMultDiv2SPK(x[xt2+1], x[xt2+0], wPiFourth)

				ur = x[xt1+0] >> 1
				ui = x[xt1+1] >> 1

				x[xt1+0] = ur + vr
				x[xt1+1] = ui - vi

				x[xt2+0] = ur - vr
				x[xt2+1] = ui + vi

				xt1 = xt2 + mh
				r -= mh << 1
				if r == 0 {
					break
				}
			}
		}

		ldm--
		if ldm == 0 {
			break
		}
	}
}
