# frozen_string_literal: true

module RecordsHelper
  def rating_state_icon(state, options)
    case state
    when "bad"
      icon "thumbs-o-down", options
    when "average"
      icon "meh-o", options
    when "good"
      icon "thumbs-o-up", options
    when "great"
      icon "heart-o", options
    end
  end
end
