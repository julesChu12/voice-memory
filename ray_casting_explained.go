package main

import (
	"fmt"
)

// Point 坐标点
type Point struct {
	X float64
	Y float64
}

// RayCasting 射线法判断点是否在多边形内
func RayCasting(point Point, polygon []Point) bool {
	n := len(polygon)
	if n < 3 {
		return false
	}

	inside := false
	j := n - 1

	for i := 0; i < n; i++ {
		// ============================================================
		// 条件1: 边的两端点必须在射线两侧（一个在上，一个在下）
		// ============================================================
		// 使用 > 而非 >= 避免射线与顶点重合时的边界问题
		oneAbove := polygon[i].Y > point.Y
		otherAbove := polygon[j].Y > point.Y

		if oneAbove != otherAbove {
			// ============================================================
			// 条件2: 计算边与射线的交点X坐标，判断是否在点右侧
			// ============================================================
			// 直线方程推导:
			// 已知: A(x₁,y₁), B(x₂,y₂), 射线y = point.Y
			// 交点: x = (x₂-x₁)*(y-y₁)/(y₂-y₁) + x₁

			intersectX := (polygon[j].X-polygon[i].X)*(point.Y-polygon[i].Y)/
				(polygon[j].Y-polygon[i].Y) + polygon[i].X

			// 如果交点在点右侧，则射线穿过这条边
			if point.X < intersectX {
				inside = !inside // 翻转状态: 奇数次=内部，偶数次=外部
			}
		}

		j = i // 移动到下一条边
	}

	return inside
}

func main() {
	// 定义一个不规则多边形（五角星形状）
	polygon := []Point{
		{3, 0},   // 顶点
		{1, 1},   // 左上
		{0, 4},   // 左顶点
		{1, 3},   // 左下
		{0, 5},   // 底部
		{2, 3},   // 内点1
		{5, 5},   // 右底
		{3, 3},   // 内点2
		{5, 2},   // 右下
		{4, 1},   // 右上
	}

	testCases := []struct {
		point Point
		desc  string
	}{
		{Point{2, 2}, "中心点（内部）"},
		{Point{3, 1}, "顶点上（边界）"},
		{Point{-1, 2}, "左侧外部"},
		{Point{6, 3}, "右侧外部"},
		{Point{2, 5}, "底部外部"},
		{Point{2.5, 3}, "凹陷区域（外部）"},
	}

	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║      射线法 (Ray Casting) 演示         ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()

	for _, tc := range testCases {
		inside := RayCasting(tc.point, polygon)
		status := "❌ 外部"
		if inside {
			status = "✅ 内部"
		}

		fmt.Printf("测试: %s\n", tc.desc)
		fmt.Printf("  点: (%.1f, %.1f)\n", tc.point.X, tc.point.Y)
		fmt.Printf("  结果: %s\n", status)

		// 可视化射线
		fmt.Printf("  射线: ─────●───────→\n")
		fmt.Println()
	}
}
