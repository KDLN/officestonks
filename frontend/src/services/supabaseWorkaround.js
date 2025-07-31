// Workaround for Supabase with noop mailer
// This allows users to register and immediately use the Office Stonks backend

import { supabase } from './supabase';
import { register as officeRegister, login as officeLogin } from './auth';

export const registerWithWorkaround = async (email, password, username) => {
  try {
    // First, try to register with Supabase
    const { error: supabaseError } = await supabase.auth.signUp({
      email,
      password,
      options: {
        data: {
          username: username || email,
        }
      }
    });

    if (supabaseError) {
      console.error('Supabase registration error:', supabaseError);
      // If Supabase fails, fall back to Office Stonks auth
      return await officeRegister(email, password);
    }

    // Since we're using noop mailer, the email won't be confirmed
    // Create a user in Office Stonks backend directly
    try {
      const officeResult = await officeRegister(email, password);
      return officeResult;
    } catch (officeError) {
      // If user already exists in Office Stonks (from previous attempt), try to login
      if (officeError.message && officeError.message.includes('already exists')) {
        return await officeLogin(email, password);
      }
      throw officeError;
    }
  } catch (error) {
    console.error('Registration workaround error:', error);
    throw error;
  }
};

export const loginWithWorkaround = async (email, password) => {
  try {
    // First try Supabase login
    const { data, error } = await supabase.auth.signInWithPassword({
      email,
      password
    });

    if (error) {
      // If email not confirmed error, try Office Stonks login
      if (error.message === 'Email not confirmed') {
        console.log('Email not confirmed in Supabase, trying Office Stonks auth');
        return await officeLogin(email, password);
      }
      throw error;
    }

    // If Supabase login works, we're good
    return { user: data.user, session: data.session };
  } catch (error) {
    console.error('Login workaround error:', error);
    // Final fallback to Office Stonks auth
    return await officeLogin(email, password);
  }
};