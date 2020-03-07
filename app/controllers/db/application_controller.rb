# frozen_string_literal: true

module DB
  class ApplicationController < ::ApplicationController
    include Pundit

    include RavenContext
    include Loggable
    include Localizable

    before_action :set_raven_context
    around_action :set_locale

    rescue_from Pundit::NotAuthorizedError, with: :user_not_authorized

    helper_method :locale_ja?, :locale_en?

    private

    def user_not_authorized
      flash[:alert] = t "messages._common.you_can_not_access_there"
      redirect_to request.referrer || db_root_path
    end
  end
end
