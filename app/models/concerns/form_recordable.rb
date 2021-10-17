# frozen_string_literal: true

module FormRecordable
  extend ActiveSupport::Concern

  included do
    include ActiveModel::Validations::Callbacks

    attr_accessor :user
    attr_reader :watched_at

    validates :user, presence: true
    validates :watched_at, presence: true
    validate :watched_at_cannot_be_in_the_future

    before_validation :set_watched_at

    def watched_at=(value)
      Time.use_zone(user.time_zone) do
        @watched_at = local_time(value)
      end
    end

    def create_activity?
      @local_watched_at.blank?
    end

    private

    def watched_at_cannot_be_in_the_future
      Time.use_zone(user.time_zone) do
        if watched_at.present? && watched_at > Time.zone.now
          i18n_path = "activemodel.errors.forms.recordable.watched_at_cannot_be_in_the_future"
          errors.add(:watched_at, I18n.t(i18n_path))
        end
      end
    end

    def set_watched_at
      Time.use_zone(user.time_zone) do
        if user.supporter?
          @local_watched_at = local_time(watched_at)
          @watched_at = persisted? ? (@local_watched_at.presence || record.watched_at) : (@local_watched_at.presence || Time.zone.now)
        else
          @watched_at = persisted? ? record.watched_at : Time.zone.now
        end
      end
    end

    def local_time(value)
      return if value.nil?

      value.is_a?(Time) ? value.in_time_zone : Time.zone.parse(value)
    end
  end
end
