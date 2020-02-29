# frozen_string_literal: true

module Db
  class ApplicationController < ::ApplicationController
    include Pundit

    include RavenContext
    include Loggable
    include RequestLocalizable

    before_action :set_raven_context
    around_action :set_locale

    rescue_from Pundit::NotAuthorizedError, with: :user_not_authorized

    private

    def user_not_authorized
      flash[:alert] = t "messages._common.you_can_not_access_there"
      redirect_to request.referrer || db_root_path
    end
  end
end
