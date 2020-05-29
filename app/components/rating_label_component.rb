# frozen_string_literal: true

class RatingLabelComponent < ApplicationComponent
  def initialize(kind:, class_name: "")
    @kind = kind.downcase.to_sym
    @class_name = class_name
  end

  private

  attr_reader :kind, :class_name

  def label_class_name
    classes = %w(badge)
    classes += class_name.split(" ")
    classes << badge_class_name
    classes.join(" ")
  end

  def icon_class_name
    classes = %w(far)
    classes << {
      great: 'fa-heart',
      good: 'fa-thumbs-up',
      average: 'fa-meh',
      bad: 'fa-thumbs-down',
    }[kind]
    classes.join(" ")
  end

  def badge_class_name
    "u-badge-#{kind}"
  end

  def label_text
    t "enumerize.episode_record.rating_state.#{kind}"
  end
end
