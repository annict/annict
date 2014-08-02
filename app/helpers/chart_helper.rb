module ChartHelper
  def chart_step_width(values)
    max = values.max
    # 一の位の数字
    ones_digit = max.to_s[-1].to_i
    # 一の位を取り除いた数字
    digit_str = max.to_s[0..-2].presence || '1'

    digit = (digit_str + '0').to_i
    digit += 10 if max >= 10 && ones_digit != 0

    digit / 10
  end
end