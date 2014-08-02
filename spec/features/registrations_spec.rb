require 'spec_helper'

describe 'ユーザ登録機能' do
  before do
    mock_auth_hash
    click_signin_with_twitter_link
  end

  context 'フォームを全て入力したとき' do
    let(:username) { SecureRandom.hex[0..10] }

    before do
      within('#new_user') do
        fill_in 'user_username', with: username
        fill_in 'user_email',    with: 'test@example.com'
        check   'user_terms'
        click_button '登録する'
      end
    end

    it 'ユーザ情報がデータベースに保存される' do
      expect(User.count).to eq 1
      expect(User.first.username).to eq username
    end
  end
end
