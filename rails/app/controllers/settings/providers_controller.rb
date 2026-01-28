# typed: false
# frozen_string_literal: true

module Settings
  class ProvidersController < ApplicationV6Controller
    before_action :authenticate_user!

    def index
    end
  end
end
