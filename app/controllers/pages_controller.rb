# frozen_string_literal: true

class PagesController < ApplicationController
  def terms
    render layout: "v1/application"
  end

  def privacy
    render layout: "v1/application"
  end
end
