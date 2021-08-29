# frozen_string_literal: true

module Forms
  class WorkRecordForm < Forms::ApplicationForm
    include Forms::Recordable

    attr_accessor :work
    attr_reader :animation_rating, :character_rating, :music_rating, :story_rating

    validates :animation_rating, allow_nil: true, inclusion: {in: Record::RATING_KINDS.map(&:to_s)}
    validates :character_rating, allow_nil: true, inclusion: {in: Record::RATING_KINDS.map(&:to_s)}
    validates :music_rating, allow_nil: true, inclusion: {in: Record::RATING_KINDS.map(&:to_s)}
    validates :story_rating, allow_nil: true, inclusion: {in: Record::RATING_KINDS.map(&:to_s)}
    validates :work, presence: true

    def animation_rating=(value)
      @animation_rating = value&.downcase.presence
    end

    def character_rating=(value)
      @character_rating = value&.downcase.presence
    end

    def music_rating=(value)
      @music_rating = value&.downcase.presence
    end

    def story_rating=(value)
      @story_rating = value&.downcase.presence
    end

    def deprecated_title=(value)
      @deprecated_title = value&.strip
    end

    def deprecated_title
      @deprecated_title.presence || ""
    end

    def body
      value = @body.presence || ""

      deprecated_title.present? ? "#{deprecated_title}\n\n#{value}" : value
    end
  end
end
