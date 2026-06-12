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
	wPiFourth := FixpSPK{Re: FixpSGL(0x5a82), Im: FixpSGL(0x5a82)}

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
