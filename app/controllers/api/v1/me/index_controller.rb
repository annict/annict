# frozen_string_literal: true

module API
  module V1
    module Me
      class IndexController < API::V1::ApplicationController
        before_action :prepare_params!, only: %i(show)

        def show
          @user = current_user
        end
      end
    end
  end
end
