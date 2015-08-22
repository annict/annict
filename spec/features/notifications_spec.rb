require 'spec_helper'

describe '通知ページ' do
  context 'ユーザにフォローされたとき' do
    let(:user1) { create(:registered_user) }
    let(:user2) { create(:registered_user) }

    before do
      user2.follow(user1)
      login_as(user1, scope: :user)

      visit '/notifications'
    end

    it 'フォローされたことが表示されること' do
      expect(page).to have_link(user2.profile.name)
    end
  end

  context "自分の記録が「いいね！」されたとき" do
    let(:user1) { create(:registered_user) }
    let(:user2) { create(:registered_user) }
    let(:checkin) { create(:checkin, user: user1) }

    before do
      user2.like_r(checkin)
      login_as(user1, scope: :user)

      visit '/notifications'
    end

    it 'いいね！されたことが表示されること' do
      path = work_episode_checkin_path(checkin.work, checkin.episode, checkin)
      expect(page).to have_link("記録", href: path)
    end

    context '「いいね！」が取り消されたとき' do
      before do
        user2.unlike_r(checkin)

        visit '/notifications'
      end

      it 'いいね！された通知が消えること' do
        expect(page).to_not have_content('「いいね！」しました')
      end
    end
  end
end
