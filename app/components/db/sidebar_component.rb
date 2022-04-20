# frozen_string_literal: true

module Db
  class SidebarComponent < Db::ApplicationComponent
    def initialize(search:)
      @search = search
    end

    private

    attr_reader :search
  end
end
