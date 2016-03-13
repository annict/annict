# frozen_string_literal: true

describe "ユーザ登録機能" do
  let!(:work) { create(:work, :with_item, :with_current_season) }
  let!(:mock_hash) { mock_auth_hash }

  before do
    click_signin_with_twitter_link
  end

  context "フォームを全て入力したとき" do
    let(:username) { SecureRandom.hex[0..10] }

    before do
      within("#new_user") do
        fill_in "user_username", with: username
        fill_in "user_email",    with: "test@example.com"
        click_button "登録する"
      end
    end

    it "ユーザ情報がデータベースに保存される" do
      expect(User.count).to eq 1
      expect(User.first.username).to eq username
    end

    it "OAuth情報がデータベースに保存される" do
      expect(Provider.first.token).to        eq mock_hash[:credentials][:token]
      expect(Provider.first.token_secret).to eq mock_hash[:credentials][:secret]
    end
  end
end
