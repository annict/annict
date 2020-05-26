# frozen_string_literal: true

module V4
  module UserDataFetchable
    extend ActiveSupport::Concern

    included do
      helper_method :user_data_fetcher_params
    end

    private

    def user_data_fetcher_params
      @user_data_fetcher_params ||= {}
    end

    def set_user_data_fetcher_params(params)
      @user_data_fetcher_params = params
    end
  end
end
