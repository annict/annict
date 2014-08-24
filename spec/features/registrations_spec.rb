require 'spec_helper'

describe 'ユーザ登録機能' do
  let(:cover_work_id) { eval(ENV['ANNICT_COVER_IMAGE_DATA']).first['work_id'] }
  let!(:work)         { create(:work, :with_item, id: cover_work_id) }
  let!(:mock_hash)    { mock_auth_hash }

  before do

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

    it 'OAuth情報がデータベースに保存される' do
      expect(Provider.first.token).to        eq mock_hash[:credentials][:token]
      expect(Provider.first.token_secret).to eq mock_hash[:credentials][:secret]
    end
  end
end
