# typed: false
# frozen_string_literal: true

describe "GET /@:username/:status_kind", type: :request do
  let!(:user) { create(:registered_user) }

  before do
    host! ENV.fetch("ANNICT_HOST")
  end

  it "アクセスできること" do
    get "/@#{user.username}/watching"

    expect(response.status).to eq(200)
    expect(response.body).to include("見てる")
  end
end
