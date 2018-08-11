# frozen_string_literal: true

module Api
  module Internal
    class AppDataController < Api::Internal::ApplicationController
      def index
        @user = current_user
        @flash = Flash.find_by(client_uuid: viewer_uuid)
        Flash.reset_data(viewer_uuid)
      end
    end
  end
end
