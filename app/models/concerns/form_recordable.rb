# frozen_string_literal: true

module FormRecordable
  extend ActiveSupport::Concern

  included do
    attr_accessor :user
    attr_reader :watched_at

    validate :watched_at_can_only_set_supporter
    validate :watched_at_cannot_be_in_the_future

    def watched_at=(value)
      return if value.nil?

      Time.use_zone(user.time_zone) do
        @watched_at = value.is_a?(Time) ? value.in_time_zone : Time.zone.parse(value)
      end
    end

    private

    def watched_at_can_only_set_supporter
      if watched_at.present? && !user.supporter?
        i18n_path = "activemodel.errors.forms.recordable.watched_at_can_only_set_supporter"
        errors.add(:watched_at, I18n.t(i18n_path))
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
