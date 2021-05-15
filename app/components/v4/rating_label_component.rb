# frozen_string_literal: true

module V4
  class RatingLabelComponent < V4::ApplicationComponent
    def initialize(rating:, advanced_rating: nil, class_name: "")
      @rating = rating.downcase.to_sym
      @advanced_rating = advanced_rating
      @class_name = class_name
    end

    private

    def label_class_name
      classes = %w[badge]
      classes += @class_name.split(" ")
      classes << badge_class_name
      classes.join(" ")
    end

    def icon_class_name
      classes = %w[far]
      classes << {
        great: "fa-heart",
        good: "fa-thumbs-up",
        average: "fa-meh",
        bad: "fa-thumbs-down"
      }[@rating]
      classes.join(" ")
    end

    def badge_class_name
      "u-badge-#{@rating}"
    end

    def label_text
      text = t "enumerize.episode_record.rating_state.#{@rating}"

      if @advanced_rating
        return "#{text} (#{@advanced_rating})"
      end

      text
    end
  end
end
