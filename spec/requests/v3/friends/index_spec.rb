# frozen_string_literal: true

describe "GET /friends", type: :request do
  let!(:user) { create(:registered_user) }

  before do
    login_as(user, scope: :user)
  end

  it "アクセスできること" do
    get "/friends"

    expect(response.status).to eq(200)
    expect(response.body).to include("SNSの友達")
  end
end
