# frozen_string_literal: true
# typed: false

require "rails_helper"

describe "Top page as guest" do
  it "displays welcome message" do
    visit "/"

    expect(page).to have_content "Welcome!"
  end
end
