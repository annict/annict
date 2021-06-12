# frozen_string_literal: true

module SentryLoadable
  extend ActiveSupport::Concern

  included do
    before_action :set_sentry_context
  end

  def set_sentry_context
    Sentry.set_user(id: current_user&.id)
  end
end
