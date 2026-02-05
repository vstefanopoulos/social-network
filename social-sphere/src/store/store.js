import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export const useStore = create(
  persist(
    (set) => ({
      // State
      user: null,
      loading: false,

      // Manually set user data
      setUser: (userData) => {
        set({ user: userData })
      },

      // Clear user (on logout)
      clearUser: () => {
        set({ user: null })
      },

      // Unread messages count
      unreadCount: 0,
      setUnreadCount: (count) => set({ unreadCount: count }),
      incrementUnreadCount: () => set((state) => ({ unreadCount: state.unreadCount + 1 })),
      decrementUnreadCount: () => set((state) => ({ unreadCount: Math.max(0, state.unreadCount - 1) })),

      // unread notifs
      unreadNotifs: 0,
      setUnreadNotifs: (count) => set({unreadNotifs: count}),
      incrementNotifs: () => set((state) => ({unreadNotifs: state.unreadNotifs + 1})),
      decrementNotifs: () => set((state) => ({unreadNotifs: Math.max(0, state.unreadNotifs - 1 ) })),

      // groupMsg
      hasMsg: false,
      setHasMsg: (hasMsg) => set({hasMsg: hasMsg}),
    }),
    {
      name: 'user',
      partialize: (state) => ({
        user: state.user
      }),

    }
  )
)

export const useMsgReceiver = create(
  persist(
    (set) => ({
      // State
      msgReceiver: null,
      loading: false,

      // set the user data of the user we want to send the msg
      setMsgReceiver: (receiverData) => {
        set({msgReceiver: receiverData})
      },

      // clear receiver
      clearMsgReceiver: () => {
        set({msgReceiver: null})
      },

    }),
    {
      name: 'msgReceiver',
      partialize: (state) => ({
        msgReceiver: state.msgReceiver
      }),
    }
  )
)