# typed: false
# frozen_string_literal: true

module Userland
  class HomeController < Userland::ApplicationController
    def show
      @categories = UserlandCategory.all.order(:sort_number)
    end
  end
end
