# frozen_string_literal: true

module Updaters
  class RecordUpdater
    attr_reader :record

    def initialize(user:, form:)
      @user = user
      @form = form
      @record = @form.record
    end

    def call
      @record.rating = @form.rating
      @record.body = @form.body
      @record.oauth_application = @form.oauth_application
      @record.detect_locale!(:body)
      @record.modified_at = Time.zone.now

      ActiveRecord::Base.transaction do
        @record.save!
        @user.touch(:record_cache_expired_at)

        if @form.share_to_twitter
          @user.share_record_to_twitter(@record)
        end
      end

      self
    end
  end
end
