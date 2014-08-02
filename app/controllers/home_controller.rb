class HomeController < ApplicationController
  layout :set_layout

  def index
    render user_signed_in? ? 'index' : 'index_guest'
  end


  private

  def set_layout
    if user_signed_in?
      'application'
    else
      'application_no_navbar'
    end
  end
end