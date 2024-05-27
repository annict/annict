# typed: false
# frozen_string_literal: true

describe "PATCH /settings/options", type: :request do
  let!(:user) { create(:registered_user) }

  before do
    login_as(user, scope: :user)
  end

  it "オプションが更新できること" do
    expect(user.hide_record_body?).to eq(true)

    patch "/settings/options", params: {setting: {hide_record_body: false}}

    expect(response.status).to eq(302)
    expect(flash[:notice]).to eq("更新しました")

    expect(user.hide_record_body?).to eq(false)
  end
end
