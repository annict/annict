# frozen_string_literal: true

module V6::SentryLoadable
  extend ActiveSupport::Concern

  def set_sentry_context
    Sentry.set_user(id: current_user&.id)
  end
end
