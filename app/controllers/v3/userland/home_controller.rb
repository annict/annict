# frozen_string_literal: true

module V3::Userland
  class HomeController < V3::Userland::ApplicationController
    def show
      @categories = UserlandCategory.all.order(:sort_number)
    end
  end
end
