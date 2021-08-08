# frozen_string_literal: true

module Chat
  class ApplicationController < ActionController::Base
    include ControllerCommon
    include Localizable

    helper_method :locale_ja?, :locale_en?, :local_url
  end
end
