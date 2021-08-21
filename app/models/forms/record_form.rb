# frozen_string_literal: true

module Forms
  class RecordForm < Forms::ApplicationForm
    attr_accessor :advanced_rating, :episode_id, :oauth_application, :record, :work_id
    attr_reader :instant, :rating, :share_to_twitter

    validates :advanced_rating, allow_nil: true, numericality: {greater_than_or_equal_to: 1, less_than_or_equal_to: 5}
    validates :body, length: {maximum: 1_048_596}
    validates :episode, presence: true, unless: :work
    validates :rating, allow_nil: true, inclusion: {in: Record::RATING_KINDS.map(&:to_s)}
    validates :work, presence: true, unless: :episode

    def body=(value)
      @body = value&.strip
    end

    def rating=(rating)
      @rating = rating&.downcase.presence
    end

    def share_to_twitter=(value)
      @share_to_twitter = ActiveModel::Type::Boolean.new.cast(value)
    end

    def instant=(value)
      @instant = ActiveModel::Type::Boolean.new.cast(value)
    end

    def episode
      @episode ||= Episode.only_kept.find_by(id: episode_id)
    end

    def work
      @work ||= episode ? episode.work : Work.only_kept.find_by(id: work_id)
    end

    def body
      @body.presence || ""
    end

    # @overload
    def persisted?
      !record.nil?
    end

    def unique_id
      @unique_id ||= SecureRandom.uuid
    end
  end
end
