# frozen_string_literal: true

module V4::Db
  class SidebarComponent < V4::Db::ApplicationComponent
    include AssetsHelper

    def initialize(search:)
      @search = search
    end

    private

    attr_reader :search
  end
end
