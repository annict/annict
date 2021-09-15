# frozen_string_literal: true

module Forms::Recordable
  extend ActiveSupport::Concern

  included do
    attr_accessor :oauth_application, :record, :user
    attr_reader :instant, :rating, :share_to_twitter, :skip_to_share, :watched_at

    validates :body, length: {maximum: 1_048_596}
    validates :rating, allow_nil: true, inclusion: {in: Record::RATING_KINDS.map(&:to_s)}
    validates :user, presence: true
    validate :watched_at_can_only_set_supporter
    validate :watched_at_cannot_be_in_the_future

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

    def skip_to_share=(value)
      @skip_to_share = ActiveModel::Type::Boolean.new.cast(value)
    end

    def watched_at=(value)
      return if value.nil?

      Time.use_zone(user.time_zone) do
        @watched_at = value.is_a?(Time) ? value.in_time_zone : Time.zone.parse(value)
      end
    end

    # @overload
    def persisted?
      !record.nil?
    end

    def unique_id
      @unique_id ||= SecureRandom.uuid
    end

    private

    def watched_at_can_only_set_supporter
      Time.use_zone(user.time_zone) do
        if watched_at.present? && !user.supporter? && watched_at != record&.watched_at&.in_time_zone
          i18n_path = "activemodel.errors.forms.recordable.watched_at_can_only_set_supporter"
          errors.add(:watched_at, I18n.t(i18n_path))
        end
      end
    end

    def watched_at_cannot_be_in_the_future
      Time.use_zone(user.time_zone) do
        if watched_at.present? && watched_at > Time.zone.now
          i18n_path = "activemodel.errors.forms.recordable.watched_at_cannot_be_in_the_future"
          errors.add(:watched_at, I18n.t(i18n_path))
        end
      end
    end
  end
end
