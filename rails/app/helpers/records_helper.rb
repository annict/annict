# typed: false
# frozen_string_literal: true

module RecordsHelper
  def rating_state_icon(state, options)
    case state
    when "bad"
      icon "thumbs-down", "far", options
    when "average"
      icon "meh", "far", options
    when "good"
      icon "thumbs-up", "far", options
    when "great"
      icon "heart", "far", options
    end
  end
end
