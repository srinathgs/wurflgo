package levenshtein

func LD(sRow, sCol string) int{
    RowLen := len(sRow)
    ColLen := len(sCol)
    var RowIdx int
    var ColIdx int
    var cost int
    if RowLen == 0 {
        return ColLen
    }
    if ColLen == 0 {
        return RowLen
    }

    v0 := make([]int, RowLen + 1)
    v1 := make([]int, RowLen + 1)
    var vTmp []int

    for RowIdx = 1; RowIdx <= RowLen; RowIdx++ {
        v0[RowIdx] = RowIdx
    }
    for ColIdx = 1; ColIdx <= ColLen; ColIdx++ {
        v1[0] = ColIdx
        ColJ := sCol[ColIdx - 1]
        for RowIdx = 1; RowIdx <= RowLen; RowIdx++ {
            RowI := sRow[RowIdx - 1]
            if ColJ == RowI {
                cost = 0
            } else {
                cost = 1
            }
            m_min := v0[RowIdx] + 1
            b := v1[RowIdx - 1] + 1
            c := v0[RowIdx - 1] + cost
            if (b < m_min){
                m_min = b;
            }
            if (c < m_min){
                m_min = c;
            }
            v1[RowIdx] = m_min;
        }
        vTmp = v0
        v0 = v1
        v1 = vTmp
    }
    return v0[RowLen]
}
