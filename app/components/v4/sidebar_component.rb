# frozen_string_literal: true

module V4
  class SidebarComponent < V4::ApplicationComponent
    def initialize(user:, search:)
      @user = user
      @search = search
    end

    private

    attr_reader :user, :search
  end
end
