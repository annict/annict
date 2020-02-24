# frozen_string_literal: true

module RavenContext
  extend ActiveSupport::Concern

  def set_raven_context
    Raven.user_context(id: current_user&.id)
    Raven.extra_context(params: params.to_unsafe_h, url: request.url)
  end
end
