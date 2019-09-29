# frozen_string_literal: true

module V3
  class ApplicationController < ActionController::Base
    include Localable

    layout "v3"

    helper_method :locale_ja?, :locale_en?, :local_url, :local_current_url, :domain_jp?
  end
end
