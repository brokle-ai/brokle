export type RowHeight = 'small' | 'medium' | 'large'

export interface CellProps {
  rowHeight?: RowHeight
  className?: string
}

export const ROW_HEIGHT_VALUES: Record<RowHeight, number> = {
  small: 40,
  medium: 60,
  large: 200,
}

export const ROW_HEIGHT_LABELS: Record<RowHeight, string> = {
  small: 'S',
  medium: 'M',
  large: 'L',
}
