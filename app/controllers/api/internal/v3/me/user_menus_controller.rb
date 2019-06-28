# frozen_string_literal: true

module Api
  module Internal
    module V3
      module Me
        class UserMenusController < Api::Internal::V3::ApplicationController
          before_action :authenticate_user!

          def show; end
        end
      end
    end
  end
end
