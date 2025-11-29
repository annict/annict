# typed: false

module Db::ApplicationHelper
  def select_diff_by_field(diffs, field)
    diffs.select { |diff| diff[1] == field.to_s }.first
  end
end
