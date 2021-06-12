# frozen_string_literal: true

module Fragment
  class ApplicationController < ActionController::Base
    include V6::PageCategorizable

    layout false
  end
end
