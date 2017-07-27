# frozen_string_literal: true

class FaqsController < ApplicationController
  def index
    @faq_categories = FaqCategory.published.order(:sort_number)
  end
end
