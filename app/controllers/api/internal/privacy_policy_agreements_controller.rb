# frozen_string_literal: true

module API
  module Internal
    class PrivacyPolicyAgreementsController < API::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        current_user.setting.update_column(:privacy_policy_agreed, true)
        head 200
      end
    end
  end
end
