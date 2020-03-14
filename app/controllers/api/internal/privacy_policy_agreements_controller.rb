# frozen_string_literal: true

module Api
  module Internal
    class PrivacyPolicyAgreementsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        current_user.setting.update_column(:privacy_policy_agreed, true)
        head 200
      end
    end
  end
end
