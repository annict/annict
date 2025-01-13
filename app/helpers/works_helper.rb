# typed: false

module WorksHelper
  def shirobako_color(round)
    return "shirobako-#{round}" if round >= 1 && round <= 6
    ""
  end
end
