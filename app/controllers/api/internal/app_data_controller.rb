# frozen_string_literal: true

module Api
  module Internal
    class AppDataController < Api::Internal::ApplicationController
      def index
        @user = current_user
      end
    end
  end
end
