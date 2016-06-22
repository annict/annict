# frozen_string_literal: true

module ItemCommon
  extend ActiveSupport::Concern

  DIFF_FIELDS = %i(name url tombo_image).freeze
  PUBLISH_FIELDS = DIFF_FIELDS + %i(work_id)

  included do
    has_attached_file :tombo_image, preserve_files: true

    validates :name, presence: true
    validates :url, presence: true, url: true, amazon: true, length: { maximum: 500 }
    validates :tombo_image, attachment_presence: true,
                            attachment_content_type: { content_type: /\Aimage/ },
                            attachment_square: true

    def to_diffable_hash
      data = self.class::DIFF_FIELDS.each_with_object({}) do |field, hash|
        hash[field] = case field
        when :tombo_image
          send(field).size
        else
          send(field)
        end

        hash
      end

      data.delete_if { |_, v| v.blank? }
    end

    def image_colors
      return @image_colors if @image_colors.present?

      image_url = decorate.image_url(:tombo_image, size: "250x250")
      colors = Miro::DominantColors.new(image_url)
      hexes = colors.to_hex
      colors = hexes.map { |hex| Paleta::Color.new(:hex, hex) }
      hexes = colors.sort_by(&:lightness).map(&:hex).map(&:downcase)

      @image_colors = {
        light: "##{hexes.last}",
        dark: "##{hexes.first}"
      }

      @image_colors
    end
  end
end
