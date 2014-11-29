require 'spec_helper'

describe '友達ページ' do
  let(:user) { create(:registered_user) }

  before do
    login_as(user, scope: :user)
  end

  context '友達がいないとき' do
    before do
      allow(user).to receive(:social_friends) { User.none }
      visit '/friends'
    end

    it 'ページが表示されること' do
      expect(page).to have_content('友達はいませんでした')
    end
  end
end
